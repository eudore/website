#!/bin/bash

# cd to code
appdir=$(go list -json github.com/eudore/website | grep Dir | cut -d\" -f4)
if [ -d $appdir ] ; then    
    cd $appdir
fi


# init sql
dbname="website"
dbuser="website"

echo "\c $dbname;"

> tmp.sql
for i in `find . -type f | grep -E 'go$'`
do
	sed -n '/^PostgreSQL Begin/,/PostgreSQL End$/p' $i | grep -Ev '(^PostgreSQL Begin|PostgreSQL End$)' >> tmp.sql
done

grep -v ^INSERT tmp.sql 

echo "GRANT ALL PRIVILEGES ON DATABASE $dbname to $dbuser;"
echo "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $dbuser;"
echo "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $dbuser;"
echo ""

grep ^INSERT tmp.sql | grep -v SELECT
grep ^INSERT tmp.sql | grep SELECT

rm -f tmp.sql