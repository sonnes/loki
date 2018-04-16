package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/pagalguy/loki/database"
	"gitlab.com/pagalguy/loki/models"
)

const (
	LOCAL_DB_URL string = "user=raviatluri password=psqlArmStrong5223 dbname=edgestore host=localhost sslmode=disable"
)

func TestHealthCHeck(t *testing.T) {

	Db := database.InitDB(LOCAL_DB_URL)

	defer Db.Close()

	req := httptest.NewRequest("GET", "/_ah/health", nil)

	res := httptest.NewRecorder()
	handler := CreateRouter(Db)

	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestInitEdgeEndpoint(t *testing.T) {

	Db := database.InitDB(LOCAL_DB_URL)

	defer Db.Close()

	testTableName := "test_create_edge_2"

	postBody := fmt.Sprintf(`
    {
      "edge" : {
        "name": "%s"
      }
    }
  `, testTableName)

	// defer cleanup
	defer func() {
		database.DropTable(Db, testTableName)
	}()

	req := httptest.NewRequest("POST", "/v1/edges/init", bytes.NewReader([]byte(postBody)))
	req.Header.Add("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := CreateRouter(Db)

	handler.ServeHTTP(res, req)

	// Check the status code is what we expect.
	if status := res.Code; status != http.StatusOK {
		log.Printf("%s", res.Body)
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestInitEdgeEndpoint_NameValidation(t *testing.T) {

	Db := database.InitDB(LOCAL_DB_URL)

	defer Db.Close()

	postBody := `
    {
      "edge" : {}
    }
  `

	req := httptest.NewRequest("POST", "/v1/edges/init", bytes.NewReader([]byte(postBody)))
	req.Header.Add("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := CreateRouter(Db)

	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestSaveEdgesEndpoint(t *testing.T) {

	Db := database.InitDB(LOCAL_DB_URL)

	defer Db.Close()

	testTableName := "test_save_edges"

	_ = database.CreateTable(Db, testTableName)

	// defer cleanup
	defer func() {
		database.DropTable(Db, testTableName)
	}()

	postBody := fmt.Sprintf(`
    {
      "edges" : [
        {
          "name" : "%[1]s",
          "src_id" : 1,
          "dest_id" : 2,
          "score" : 10.12,
          "data" : {
            "key1" : "value1"
          }
        },
        {
          "name" : "%[1]s",
          "src_id" : 2,
          "dest_id" : 2,
          "score" : 103125235.113352,
          "data" : {
            "key2" : "value2"
          }
        }
      ]
    }
  `, testTableName)

	req := httptest.NewRequest("POST", "/v1/edges/save", bytes.NewReader([]byte(postBody)))
	req.Header.Add("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := CreateRouter(Db)

	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestDeleteEdgesEndpoint(t *testing.T) {

	Db := database.InitDB(LOCAL_DB_URL)

	defer Db.Close()

	testTableName := "test_delete_edges"

	_ = database.CreateTable(Db, testTableName)

	// defer cleanup
	defer func() {
		database.DropTable(Db, testTableName)
	}()

	edgesList := make([]models.Edge, 2)

	edgesList[0] = models.Edge{
		Name:   testTableName,
		SrcId:  1,
		DestId: 2,
	}

	edgesList[1] = models.Edge{
		Name:   testTableName,
		SrcId:  3,
		DestId: 4,
	}

	_ = models.SaveMany(Db, &edgesList)

	postBody := fmt.Sprintf(`
    {
      "edges" : [
        {
          "name" : "%[1]s",
          "src_id" : 1,
          "dest_id" : 2
        },
        {
          "name" : "%[1]s",
          "src_id" : 3,
          "dest_id" : 4
        }
      ]
    }
  `, testTableName)

	req := httptest.NewRequest("POST", "/v1/edges/delete", bytes.NewReader([]byte(postBody)))
	req.Header.Add("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := CreateRouter(Db)

	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
