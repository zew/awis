#!/bin/sh
# use dos2unix

# wget "localhost:8081/alexa_web_information_service/top-sites?Url=www.zew.de&Start=0&Count=1&CountryCode=DE&submit=+Submit+"
crontab -l
crontab -e

#Minute  Hour  DayofMonth  Month  DayofWeek  Command
0        9     *           *      *          wget "localhost:8081/alexa_web_information_service/top-sites-auto"