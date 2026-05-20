module linkshelf

go 1.25.6

replace github.com/polecat/linkshelf/internal/store => ../internal/store

require github.com/mattn/go-sqlite3 v1.14.44
