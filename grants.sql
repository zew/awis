CREATE USER 'category-editor'@'%'         IDENTIFIED BY 'pass1';
CREATE USER 'category-editor'@'localhost' IDENTIFIED BY 'pass1';

GRANT SELECT , 
    INSERT ( category_mcafee, category_statista, category_custom) , 
    UPDATE ( category_mcafee, category_statista, category_custom)   
ON awisdb.category_zew   TO 'category-editor'@'%' IDENTIFIED BY 'pass1' ;
GRANT SELECT ON awisdb.* TO 'category-editor'@'%' IDENTIFIED BY 'pass1' ;


GRANT SELECT , 
    INSERT ( category_mcafee, category_statista, category_custom) , 
    UPDATE ( category_mcafee, category_statista, category_custom)   
ON awisdb.category_zew   TO 'category-editor'@'%' IDENTIFIED BY 'pass1' ;
GRANT SELECT ON awisdb.* TO 'category-editor'@'%' IDENTIFIED BY 'pass1' ;

SHOW WARNINGS;
FLUSH PRIVILEGES;

-- ##################################################################

CREATE USER 'category-superuser'@'%'         IDENTIFIED BY 'pass2';
CREATE USER 'category-superuser'@'localhost' IDENTIFIED BY 'pass2';

GRANT ALL PRIVILEGES ON awisdb.category_zew  TO 'category-superuser'@'%' IDENTIFIED BY 'pass2' ;
GRANT SELECT         ON awisdb.*             TO 'category-superuser'@'%' IDENTIFIED BY 'pass2' ;
GRANT EXECUTE        ON awisdb.*             TO 'category-superuser'@'%' IDENTIFIED BY 'pass2' ;


GRANT ALL PRIVILEGES ON awisdb.category_zew  TO 'category-superuser'@'%' IDENTIFIED BY 'pass2' ;
GRANT SELECT         ON awisdb.*             TO 'category-superuser'@'%' IDENTIFIED BY 'pass2' ;
GRANT EXECUTE        ON awisdb.*             TO 'category-superuser'@'%' IDENTIFIED BY 'pass2' ;

SHOW WARNINGS;
FLUSH PRIVILEGES;

-- ##################################################################

DROP PROCEDURE import_new_entries;

delimiter //

CREATE PROCEDURE import_new_entries ()
     BEGIN
        INSERT INTO category_zew  (domain_name, rank_avg, added_pageviews, category_custom)  
        SELECT t2.domain_name, t2.rank_avg, t2.added_pageviews, ''
          FROM domain_aggregated t2
        ON DUPLICATE KEY UPDATE 
            last_updated = NOW(),
            rank_avg = t2.rank_avg,
            added_pageviews = t2.added_pageviews
        ;  
     END//
delimiter ;



CREATE DEFINER=`root`@`localhost` 
SQL SECURITY DEFINER VIEW `cat_statista` AS 
    SELECT category_statista, count(*) anz, sum(added_pageviews) added_pv  FROM 
    `category_zew` 
    WHERE 1
    group by `category_statista`     

;

