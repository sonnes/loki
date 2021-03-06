package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/sonnes/loki/ds_to_sql"
)

/*
 Edge CSV
 name
 src_id
 src_type
 dest_id
 dest_type
 score
 status
 updated
 data
*/

type Blank struct {
	Key *ds_to_sql.Key
}

func (dsObj *Blank) SetKey(entityKey *ds_to_sql.Key) {
	dsObj.Key = entityKey
}

func (dsObj *Blank) ExtractCSV() *[][]string {

	return &[][]string{
		[]string{},
	}

}

type Entity struct {
	Key         *ds_to_sql.Key
	Name        string           `datastore:"name"`
	ParentKey   *ds_to_sql.Key   `datastore:"parent_key"`
	EntityKeys  []*ds_to_sql.Key `datastore:"auto_entity_keys"`
	Updated     *time.Time       `datastore:"updated"`
	OrderNo     int              `datastore:"score"`
	Noun        string           `datastore:"noun"`
	ContentType string           `datastore:"content_type"`
}

func (dsObj *Entity) SetKey(entityKey *ds_to_sql.Key) {
	dsObj.Key = entityKey
}

func (dsObj *Entity) ExtractCSV() *[][]string {

	allEdges := make([][]string, 0)

	score := "null"
	if dsObj.Updated != nil {
		score = dsObj.Updated.UTC().Format(time.RFC3339)
	} else {
		score = strconv.Itoa(dsObj.OrderNo)
	}

	data := "null"
	if dsObj.Noun != "" {
		dataJson := map[string]string{
			"noun": dsObj.Noun,
		}
		dataBytes, _ := json.Marshal(dataJson)
		data = string(dataBytes)
	}

	if dsObj.ParentKey != nil {

		srcId := strconv.FormatInt(dsObj.Key.IntID(), 10)
		destId := strconv.FormatInt(dsObj.ParentKey.IntID(), 10)
		id := srcId + ":" + destId

		parentCSV := []string{
			"child_page",
			id,
			// src
			srcId,
			strings.ToLower(dsObj.Key.Kind()),
			// dest
			destId,
			strings.ToLower(dsObj.ParentKey.Kind()),
			score,
			"active",
			data,
		}

		allEdges = append(allEdges, parentCSV)
	}

	if dsObj.EntityKeys != nil {
		for _, entityKey := range dsObj.EntityKeys {

			srcId := strconv.FormatInt(dsObj.Key.IntID(), 10)
			destId := strconv.FormatInt(entityKey.IntID(), 10)
			id := srcId + ":" + destId

			tagCSV := []string{
				"tag",
				id,
				// src
				srcId,
				strings.ToLower(dsObj.Key.Kind()),
				// dest
				destId,
				strings.ToLower(entityKey.Kind()),
				score,
				"active",
				data,
			}
			allEdges = append(allEdges, tagCSV)
		}
	}

	return &allEdges

}

type Post struct {
}

type Follow struct {
	Key      *ds_to_sql.Key
	SrcKey   *ds_to_sql.Key `datastore:"src_key"`
	DestKey  *ds_to_sql.Key `datastore:"dest_key"`
	DestType string         `datastore:"dest_type"`
	Created  time.Time      `datastore:"created"`
	Updated  time.Time      `datastore:"updated"`
	Source   string         `datastore:"source"`
}

func (dsObj *Follow) SetKey(entityKey *ds_to_sql.Key) {
	dsObj.Key = entityKey
}

func (dsObj *Follow) ExtractCSV() *[][]string {

	score := "null"
	updated := "null"
	if !dsObj.Updated.IsZero() {
		updated = dsObj.Updated.UTC().Format(time.RFC3339)
		score = strconv.FormatInt(UnixMilli(&dsObj.Updated), 10)
	} else if !dsObj.Created.IsZero() {
		updated = dsObj.Created.UTC().Format(time.RFC3339)
		score = strconv.FormatInt(UnixMilli(&dsObj.Created), 10)
	}

	srcId := strconv.FormatInt(dsObj.SrcKey.IntID(), 10)
	destId := strconv.FormatInt(dsObj.DestKey.IntID(), 10)
	id := srcId + ":" + destId

	data := "null"
	if dsObj.Source != "" {
		dataJson := map[string]string{
			"source": dsObj.Source,
		}
		dataBytes, _ := json.Marshal(dataJson)
		data = string(dataBytes)
	}

	edgeCSV := []string{
		"follow",
		id,
		srcId,
		strings.ToLower(dsObj.SrcKey.Kind()),
		destId,
		strings.ToLower(dsObj.DestKey.Kind()),
		score,
		"active",
		updated,
		data,
	}

	return &[][]string{edgeCSV}
}

func UnixMilli(ts *time.Time) int64 {
	return ts.UnixNano() / int64(time.Millisecond)
}
