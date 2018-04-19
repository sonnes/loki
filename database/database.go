package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/ExpansiveWorlds/instrumentedsql"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	DML_CREATE_EDGE_TABLE string = `
	  CREATE TABLE IF NOT EXISTS %s (
	    id varchar PRIMARY KEY,
	    src_id bigint,
	    src_type varchar,
	    dest_id bigint,
	    dest_type varchar,
	    score decimal,
	    data jsonb,
	    status varchar,
	    updated timestamp
	  );
	`

	DML_DEFAULT_INDEXES string = `
	  CREATE INDEX IF NOT EXISTS %[1]s_src_id ON %[1]s (src_id);
	  CREATE INDEX IF NOT EXISTS %[1]s_dest_id ON %[1]s (dest_id);
	  CREATE INDEX IF NOT EXISTS %[1]s_score ON %[1]s (score);
	  CREATE INDEX IF NOT EXISTS %[1]s_status ON %[1]s (status);
	  CREATE INDEX IF NOT EXISTS %[1]s_combi ON %[1]s (src_id, dest_id, score, status);
	`

	DML_DROP_TABLE string = "DROP TABLE %s"
)

func InitDB(databaseURL string) *sqlx.DB {

	dbDriver, err := sql.Open("postgres", databaseURL)
	Db := sqlx.NewDb(dbDriver, "postgres")

	if err != nil {
		log.Fatalf("could not open database connection: %v\n", err)
	}
	return Db
}

func InitDebugDB(databaseURL string) *sqlx.DB {

	logger := instrumentedsql.LoggerFunc(func(ctx context.Context, msg string, keyvals ...interface{}) {
		log.Printf("%s %v", msg, keyvals)
	})

	sql.Register("instrumented-postgres", instrumentedsql.WrapDriver(&pq.Driver{}, instrumentedsql.WithLogger(logger)))

	dbDriver, err := sql.Open("instrumented-postgres", databaseURL)
	Db := sqlx.NewDb(dbDriver, "postgres")

	if err != nil {
		log.Fatalf("could not open database connection: %v\n", err)
	}
	return Db
}

func CreateTable(Db *sqlx.DB, tableName string) error {

	query := fmt.Sprintf(DML_CREATE_EDGE_TABLE, tableName)

	_, err := Db.Exec(query)

	return err
}

func CreateDefaultIndexes(Db *sqlx.DB, tableName string) error {
	query := fmt.Sprintf(DML_DEFAULT_INDEXES, tableName)

	_, err := Db.Exec(query)

	return err
}

func DropTable(Db *sqlx.DB, tableName string) error {

	query := fmt.Sprintf(DML_DROP_TABLE, tableName)

	_, err := Db.Exec(query)

	return err
}

func MakeRange(min int, max int) []int {
	list := make([]int, max-min+1)
	for i := range list {
		list[i] = i + min
	}
	return list
}

func GeneratePlaceholder(start int, count int) string {
	argRange := MakeRange(start, start+count-1)

	placeholders := make([]string, len(argRange))

	for argIdx, value := range argRange {
		placeholders[argIdx] = fmt.Sprintf("$%d", value)
	}

	return "(" + strings.Join(placeholders, ",") + ")"
}
