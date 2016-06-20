// Copyright 2012 James Cooper. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package gorp provides a simple way to marshal Go structs to and from
// SQL databases.  It uses the database/sql package, and should work with any
// compliant database/sql driver.
//
// Source code and project home:
// https://github.com/go-gorp/gorp

package gorp

import (
	"database/sql"
	"fmt"
	"reflect"
)

// SelectInt executes the given query, which should be a SELECT statement for a single
// integer column, and returns the value of the first row returned.  If no rows are
// found, zero is returned.
func SelectInt(e SqlExecutor, query string, args ...interface{}) (int64, error) {
	var h int64
	err := selectVal(e, &h, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return h, nil
}

// SelectNullInt executes the given query, which should be a SELECT statement for a single
// integer column, and returns the value of the first row returned.  If no rows are
// found, the empty sql.NullInt64 value is returned.
func SelectNullInt(e SqlExecutor, query string, args ...interface{}) (sql.NullInt64, error) {
	var h sql.NullInt64
	err := selectVal(e, &h, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return h, err
	}
	return h, nil
}

// SelectFloat executes the given query, which should be a SELECT statement for a single
// float column, and returns the value of the first row returned. If no rows are
// found, zero is returned.
func SelectFloat(e SqlExecutor, query string, args ...interface{}) (float64, error) {
	var h float64
	err := selectVal(e, &h, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return h, nil
}

// SelectNullFloat executes the given query, which should be a SELECT statement for a single
// float column, and returns the value of the first row returned. If no rows are
// found, the empty sql.NullInt64 value is returned.
func SelectNullFloat(e SqlExecutor, query string, args ...interface{}) (sql.NullFloat64, error) {
	var h sql.NullFloat64
	err := selectVal(e, &h, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return h, err
	}
	return h, nil
}

// SelectStr executes the given query, which should be a SELECT statement for a single
// char/varchar column, and returns the value of the first row returned.  If no rows are
// found, an empty string is returned.
func SelectStr(e SqlExecutor, query string, args ...interface{}) (string, error) {
	var h string
	err := selectVal(e, &h, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return h, nil
}

// SelectNullStr executes the given query, which should be a SELECT
// statement for a single char/varchar column, and returns the value
// of the first row returned.  If no rows are found, the empty
// sql.NullString is returned.
func SelectNullStr(e SqlExecutor, query string, args ...interface{}) (sql.NullString, error) {
	var h sql.NullString
	err := selectVal(e, &h, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return h, err
	}
	return h, nil
}

// SelectOne executes the given query (which should be a SELECT statement)
// and binds the result to holder, which must be a pointer.
//
// If no row is found, an error (sql.ErrNoRows specifically) will be returned
//
// If more than one row is found, an error will be returned.
//
func SelectOne(m *DbMap, e SqlExecutor, holder interface{}, query string, args ...interface{}) error {
	t := reflect.TypeOf(holder)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return fmt.Errorf("gorp: SelectOne holder must be a pointer, but got: %t", holder)
	}

	// Handle pointer to pointer
	isptr := false
	if t.Kind() == reflect.Ptr {
		isptr = true
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		var nonFatalErr error

		list, err := hookedselect(m, e, holder, query, args...)
		if err != nil {
			if !NonFatalError(err) { // FIXME: double negative, rename NonFatalError to FatalError
				return err
			}
			nonFatalErr = err
		}

		dest := reflect.ValueOf(holder)
		if isptr {
			dest = dest.Elem()
		}

		if list != nil && len(list) > 0 { // FIXME: invert if/else
			// check for multiple rows
			if len(list) > 1 {
				return fmt.Errorf("gorp: multiple rows returned for: %s - %v", query, args)
			}

			// Initialize if nil
			if dest.IsNil() {
				dest.Set(reflect.New(t))
			}

			// only one row found
			src := reflect.ValueOf(list[0])
			dest.Elem().Set(src.Elem())
		} else {
			// No rows found, return a proper error.
			return sql.ErrNoRows
		}

		return nonFatalErr
	}

	return selectVal(e, holder, query, args...)
}

func selectVal(e SqlExecutor, holder interface{}, query string, args ...interface{}) error {
	if len(args) == 1 {
		switch m := e.(type) {
		case *DbMap:
			query, args = maybeExpandNamedQuery(m, query, args)
		case *Transaction:
			query, args = maybeExpandNamedQuery(m.dbmap, query, args)
		}
	}
	rows, err := e.query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}

	return rows.Scan(holder)
}

func hookedselect(m *DbMap, exec SqlExecutor, i interface{}, query string,
	args ...interface{}) ([]interface{}, error) {

	var nonFatalErr error

	list, err := rawselect(m, exec, i, query, args...)
	if err != nil {
		if !NonFatalError(err) {
			return nil, err
		}
		nonFatalErr = err
	}

	// Determine where the results are: written to i, or returned in list
	if t, _ := ToSliceType(i); t == nil {
		for _, v := range list {
			if v, ok := v.(HasPostGet); ok {
				err := v.PostGet(exec)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		resultsValue := reflect.Indirect(reflect.ValueOf(i))
		for i := 0; i < resultsValue.Len(); i++ {
			if v, ok := resultsValue.Index(i).Interface().(HasPostGet); ok {
				err := v.PostGet(exec)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return list, nonFatalErr
}

func rawselect(m *DbMap, exec SqlExecutor, i interface{}, query string,
	args ...interface{}) ([]interface{}, error) {
	var (
		appendToSlice   = false // Write results to i directly?
		intoStruct      = true  // Selecting into a struct?
		pointerElements = true  // Are the slice elements pointers (vs values)?
	)

	var nonFatalErr error

	// get type for i, verifying it's a supported destination
	t, err := ToStructType(i) // to struct type
	if err != nil {
		var err2 error
		if t, err2 = ToSliceType(i); t == nil {
			if err2 != nil {
				return nil, err2
			}
			return nil, err
		}
		pointerElements = t.Kind() == reflect.Ptr
		if pointerElements {
			t = t.Elem()
		}
		appendToSlice = true
		intoStruct = t.Kind() == reflect.Struct
	}

	// If the caller supplied a single struct/map argument, assume a "named
	// parameter" query.  Extract the named arguments from the struct/map, create
	// the flat arg slice, and rewrite the query to use the dialect's placeholder.
	if len(args) == 1 {
		query, args = maybeExpandNamedQuery(m, query, args)
	}

	// Run the query
	rows, err := exec.query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Fetch the column names as returned from db
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !intoStruct && len(cols) > 1 {
		return nil, fmt.Errorf("gorp: select into non-struct slice requires 1 column, got %d", len(cols))
	}

	var fieldsIdxChains [][][]int
	if intoStruct {
		_, fieldsIdxChains, err = colToFieldIdxChain(m, t, i, cols)
		// logl.Printf("%s", util.IndentedDump(fieldsIdxChains))
		if err != nil {
			if !NonFatalError(err) {
				return nil, err
			}
			nonFatalErr = err
			if err != nil {
				logl.Printf("SETTING MISSING COLUMN ERROR: %v", err)
			}
		}
	}

	conv := m.TypeConverter

	// Add results to one of these two slices.
	var (
		list       = make([]interface{}, 0)
		sliceValue = reflect.Indirect(reflect.ValueOf(i))
	)
	// logl.Printf("sliceValue %v - %T - %T\n", sliceValue, sliceValue, sliceValue.Interface())

	rowsCount := 0
	for {
		if !rows.Next() {
			// if error occured return rawselect
			if rows.Err() != nil {
				return nil, rows.Err()
			}
			// time to exit from outer "for" loop
			break
		}
		rowsCount++
		dest := make([]interface{}, len(cols))
		// logl.Printf("Dest numColumns: %v - %T\n", len(dest), dest)

		custScan := make([]CustomScanner, 0)

		v := reflect.New(t) // Pointer to ZeroVal
		dbg := v.Elem()
		if rowsCount < 3 {
			logl.Printf("rawSelect03 into %v %T", dbg.Kind(), dbg.Interface())
		}

		for x := range cols {
			if intoStruct {
				if false {
					// Previously
					// f = f.FieldByIndex(index)
				} else {
					fieldChains := fieldsIdxChains[x]
					if fieldChains == nil {
						// this field is not present in the struct, so create a dummy
						// value for rows.Scan to scan into
						var dummy dummyField
						dest[x] = &dummy
						continue
					}

					for _, fieldChain := range fieldsIdxChains[x] {
						f := v.Elem()
						// for _, fldIdx := range fieldInstance { // old
						// }
						func() {
							defer func() {
								if r := recover(); r != nil {
									logl.Printf("ERROR in field index chain: %v %v - %v", x, fieldsIdxChains[x], r)
								}
							}()
							f = f.FieldByIndex(fieldChain)
						}()
						target := f.Addr().Interface() // make to pointer
						if rowsCount < 2 && x < 10 {
							logl.Printf("\trawSelect11-%-2v %-24v %-14v %-14v %+v  => %T", x, cols[x], f.Kind(), f.Type(), fieldChain, target)
						}
						// Custom scanner
						if conv != nil {
							scanner, ok := conv.FromDb(target)
							if ok {
								target = scanner.Holder
								custScan = append(custScan, scanner)
							}
						}
						dest[x] = target
						// break
					}
				}
			} else {
				dest[0] = v.Elem().Addr().Interface()
				// logl.Printf("rawSelect05 - scalar %#v \n", dest)
			}
		}

		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		// logl.Printf("\trawSelect04 - %v", len(dest))

		for _, c := range custScan {
			err = c.Bind()
			if err != nil {
				return nil, err
			}
		}

		//
		//
		//
		if rowsCount < 3 {
			logl.Printf("post scan treatment of duplicate columns\n")
			// 	v.Elem().FieldByIndex([]int{0}).SetString("aaa")
		}
		for x := range cols {
			if intoStruct {
				fieldChains := fieldsIdxChains[x]
				if fieldChains == nil {
					continue
				}
				if len(fieldChains) > 1 {
					// fieldValues := []reflect.Value{}
					var nonZeroVal reflect.Value
					// var nonZeroInt int64 = 0
					found := false
					for _, fieldChain := range fieldsIdxChains[x] {
						f1 := v.Elem().FieldByIndex(fieldChain)
						zeroVal := reflect.New(f1.Type())
						zeroVal = zeroVal.Elem() // dereference it
						if rowsCount < 2 {
							logl.Printf("rawSelect13-%-2v %-24v %-14v %-14v %+v  => %v", x, cols[x], f1.Kind(), f1.Type(), fieldChain, f1.Interface())
						}
						if f1.Interface() != zeroVal.Interface() {
							nonZeroVal = f1
							found = true
							if rowsCount < 2 {
								logl.Printf("rawSelect14 nonZeroVal is now  %v (%v)", nonZeroVal.Interface(), zeroVal.Interface())
							}
						}
					}
					if found {
						for _, fieldChain := range fieldsIdxChains[x] {
							f1 := v.Elem().FieldByIndex(fieldChain)
							f1.Set(nonZeroVal)
						}
					}
				}
			}
		}

		// logl.Printf("\trawSelect 05 - %v", len(dest))

		if appendToSlice {
			if !pointerElements {
				v = v.Elem()
			}
			sliceValue.Set(reflect.Append(sliceValue, v))
		} else {
			logl.Printf("rawSelect16: selectOne; select into single struct => appending to list\n")
			list = append(list, v.Interface())
		}

		// logl.Printf("\trawSelect 06 - %v", len(list))
	}

	if appendToSlice && sliceValue.IsNil() {
		sliceValue.Set(reflect.MakeSlice(sliceValue.Type(), 0, 0))
	}

	logl.Printf("rawSelect16: nonFatal is %v", nonFatalErr)
	// logl.Printf("rawSelect07 %v - %v NONFAT: %v", sliceValue.Len(), len(list), nonFatalErr)

	return list, nonFatalErr
}
