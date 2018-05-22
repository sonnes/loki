package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.com/pagalguy/loki/database"
)

const (
	ACTIVE  string = "active"
	DELETED string = "deleted"
)

type Edge struct {
	Id       string     `json:"id,omitempty" db:"id"`
	Name     *string    `json:"name" db:"name"`
	SrcId    int64      `json:"src_id" db:"src_id"`
	SrcType  *string    `json:"src_type,omitempty" db:"src_type"`
	DestId   int64      `json:"dest_id" db:"dest_id"`
	DestType *string    `json:"dest_type,omitempty" db:"dest_type"`
	Score    float32    `json:"score,omitempty" db:"score,decimal"`
	Data     *Data      `json:"data,omitempty" db:"data"`
	Status   string     `json:"status,omitempty" db:"status"`
	Updated  *time.Time `json:"updated,omitempty" db:"updated,timestamp"`
}

type Data map[string]interface{}

func (data *Data) Value() (driver.Value, error) {

	if data != nil {
		jstr, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return jstr, nil
	}

	return nil, nil
}

func (em *Data) Scan(src interface{}) error {

	var data []byte
	if b, ok := src.([]byte); ok {
		data = b
	} else if s, ok := src.(string); ok {
		data = []byte(s)
	}
	return json.Unmarshal(data, em)
}

func (edge *Edge) DbId() string {

	id := fmt.Sprintf("%d:%d", edge.SrcId, edge.DestId)
	return id
}

const (
	INSERT_PART string = `
		INSERT INTO %s (
		  id,
		  src_id,
		  src_type,
		  dest_id,
		  dest_type,
		  score,
		  data,
		  status,
		  updated
		)
		VALUES
	`
	// The on conflict part makes sure that slate data
	// is not updated into DB.
	UPDATE_PART string = `
		ON CONFLICT (id) DO UPDATE
			SET
				src_type = CASE WHEN %[1]s.updated < EXCLUDED.updated
					THEN EXCLUDED.src_type
					ELSE %[1]s.src_type END,
				dest_type = CASE WHEN %[1]s.updated < EXCLUDED.updated
					THEN EXCLUDED.dest_type
					ELSE %[1]s.dest_type END,
				score = CASE WHEN %[1]s.updated < EXCLUDED.updated
					THEN EXCLUDED.score
					ELSE %[1]s.score END,
				data = CASE WHEN %[1]s.updated < EXCLUDED.updated
					THEN EXCLUDED.data
					ELSE %[1]s.data END,
				status = CASE WHEN %[1]s.updated < EXCLUDED.updated
					THEN EXCLUDED.status
					ELSE %[1]s.status END,
				updated = CASE WHEN %[1]s.updated < EXCLUDED.updated
					THEN EXCLUDED.updated
					ELSE %[1]s.updated END
	`
	DELETE_PART = "UPDATE %s SET status=$1, updated=$2 WHERE id IN %s"
)

func SaveMany(db *sqlx.DB, edgesPtr *[]Edge) error {

	groupedEdges := GroupByEdgeName(edgesPtr)

	for edgeName, edges := range groupedEdges {

		query := fmt.Sprintf(INSERT_PART, edgeName)

		valueStrings := make([]string, 0, len(edges))
		valueArgs := make([]interface{}, 0, len(edges)*9)

		for idx, edge := range edges {

			placeholder := database.GeneratePlaceholder(idx*9+1, 9)

			valueStrings = append(valueStrings, placeholder)
			valueArgs = append(
				valueArgs, edge.DbId(), edge.SrcId,
				edge.SrcType, edge.DestId, edge.DestType, edge.Score,
				edge.Data, edge.Status, edge.Updated,
			)
		}

		query = query + strings.Join(valueStrings, " , ")

		query = query + fmt.Sprintf(UPDATE_PART, edgeName)

		_, err := db.Exec(query, valueArgs...)

		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteMany(db *sqlx.DB, edgesPtr *[]Edge) error {

	groupedEdges := GroupByEdgeName(edgesPtr)

	for edgeName, edges := range groupedEdges {

		placeholder := database.GeneratePlaceholder(3, len(edges))

		deleteQuery := fmt.Sprintf(DELETE_PART, edgeName, placeholder)

		valueArgs := make([]interface{}, 0)

		valueArgs = append(valueArgs, DELETED, time.Now())

		for _, edge := range edges {
			valueArgs = append(valueArgs, edge.DbId())
		}

		_, err := db.Exec(deleteQuery, valueArgs...)

		if err != nil {
			return err
		}
	}

	return nil
}

func RunQuery(db *sqlx.DB, query string) (*[]Edge, error) {

	edgeList := make([]Edge, 0)

	err := db.Select(&edgeList, query)

	if err != nil {
		return nil, err
	}

	return &edgeList, nil
}

func GroupByEdgeName(edgesPtr *[]Edge) map[string][]Edge {

	allEdges := *edgesPtr

	groupedEdges := make(map[string][]Edge)

	for _, edge := range allEdges {
		groupedEdges[*edge.Name] = append(groupedEdges[*edge.Name], edge)
	}

	return groupedEdges
}
