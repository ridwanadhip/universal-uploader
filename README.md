## What is this?
Universal uploader is a script which will be used by engineers to substitute script development for performing various tasks, such as:
1. Create Jarvis script for uploading bulk data
2. Uploading bulk data via computer
3. Bulk updating data for solving data issue

Rather than create new script, we only need to create a config file. Then the universal uploader will determine the structure of the input file and the target which the data will be uploaded into it. By using this we don't need to develop script anymore for bulk uploading data. Though anything more complex still need a independent script to be performed.

Currently universal uploader only supported CSV as input and MySQL as the target. But more format will be planned in the future, such as JSON, HTTP and Redis.

Feature planned:
1. Outputting final data to a file such as CSV
2. Validation via dry run
4. Resume last failed run
5. Hook support so this script can be used as library for other programs

Please refers to example folder for complete example of usage.

## Usage
```
universal-uploader [--verbose] [--resume] [--dry-run] <config-file-path> <input-file-path>
```
- verbose: enable more logging to terminal
- resume: resume last failed run, start at the last failed line.
- dry run: (under development) do validation without inserting data to data destination

Example:
```
$ ./universal-uploader --verbose config.yaml path/data.csv
```

## Config file examples
Lets say we have 1 csv:
```
primary,col2,col3
val1,2,3
val2,5,6
```

### Example 1
Upload to a mysql table that has column with same names and has same total of columns, all in string types. This script will determine the csv column and database column based on given csv file, which are primary, col2, and col3 :
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
```


### Example 2
The csv structure is same as example 1, but the size of data is huge and we need to insert it using batch processing:
```
batchSize: 500
delay: 200 # in ms
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
```

### Example 3
We only need to insert first and third column to table, we can do it by choosing the csv columns using column name as reference id:
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
      - name: primary
        value: ^primary^
      - name: col3
        value: ^col3^
```

### Example 4
We need to do upsertion because existing data has same primary key values:
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
    upsert: true
    fields:
      - name: primary
        value: ^primary^
        uniqueValue: true
      - name: col2
        value: ^col2^
        replaceOldValue: true
      - name: col3
        value: ^col3^
        replaceOldValue: true
```

### Example 5
Inject envar to config file:
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: $USERNAME$
    password: $PASSWORD$
```

### Example 6
CSV column and SQL table has different name, use csv column reference id to do it:
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
    fields:
      - name: primaryDB
        value: ^primary^
      - name: col2DB
        value: ^col2^
      - name: col3DB
        value: ^col3^
```

### Example 7
Convert csv column type:
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
    fields:
      - name: primaryDB
        type: string
        value: ^primary^
      - name: col2DB
        type: integer
        value: ^col2^
      - name: col3DB
        type: decimal
        value: ^col2^
```

### Example 8
Complex column value formatting and hardcoded value
```
targets:
  - type: mysql
    name: databaseName
    dataName: tableName
    host: test
    username: test
    password: test
    fields:
      - name: primary
      - name: column2
        type: string
        value: "{name: '^col2^', value: 0}"
      - name: column3
        value: '' # empty string
      - name: column4
        value: NIL # null sql value
```