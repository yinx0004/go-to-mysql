# go-to-mysql

## Description
This tool will do the following things
1. create connection pool to MySQL
2. create database if not exists
3. create table `test_tab` is not exists
4. insert into `test_tab` with random values without query timeout
```
create table test_tab (
    id int not null auto_increment primary key,
    col1 int not null, 
    col2 char(20) not null
    )
```

## How to build
### Build for macOS
```
GOOS=darwin GOARCH=amd64 go build -o go-to-mysql
```

### Build for Linux
```
GOOS=linux GOARCH=amd64 go build -o go-to-mysql
```


## Usage
```
./go-to-mysql --help
  -P string
        MySQL server port (default "3306")
  -T string
        MySQL parseTime(true|false) (default "true")
  -c int
        Number of Goroutione (default 50)
  -d string
        MySQL database name
  -debug
        show debug level log
  -h string
        MySQL host (default "localhost")
  -mode string
        [HealthCheck|Write|RW] (default "HealthCheck")
  -p string
        MySQL password
  -sleep duration
        Sleep time, support time duration [s|m|h] (default 1s)
  -u string
        MySQL user (default "root")
  -version
        Display version and exit

```

For example
```
./go-to-mysql -h="202.81.116.204" -u="ops" -p="password" -d="db10" -c=1 -sleep=2s -mode="RW" -debug
```
> make sure your mysql account has create database, create table, select, insert privileges, 
> and you need `super` privilege to get active **master hostname**
