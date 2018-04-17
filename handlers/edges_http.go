package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"gitlab.com/pagalguy/loki/database"
	"gitlab.com/pagalguy/loki/models"
)

type InitEdgeRequest struct {
	Name      string
	Namespace *string
}

func InitEdgeEndpoint(Db *sqlx.DB, w http.ResponseWriter, r *http.Request) {

	var jsonBody InitEdgeRequest

	if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	defer r.Body.Close()

	if jsonBody.Name == "" {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: "Name is required to initialize the edge",
			Fields:  &[]string{"name"},
		})
		return
	}

	err := database.CreateTable(Db, jsonBody.Name)

	if err != nil {
		WriteError(w, &AppError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	err = database.CreateDefaultIndexes(Db, jsonBody.Name)

	if err != nil {
		WriteError(w, &AppError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	responseJson := make(map[string]string)

	responseJson["success"] = "true"
	responseJson["message"] = fmt.Sprintf("%s - edge has been created successfully:", jsonBody.Name)

	WriteJson(w, responseJson, http.StatusOK)
}

type EdgesListRequest struct {
	Edges *[]models.Edge
}

func SaveEdgesEndpoint(db *sqlx.DB, w http.ResponseWriter, r *http.Request) {

	var jsonBody EdgesListRequest

	if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	defer r.Body.Close()

	if len(*jsonBody.Edges) == 0 {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: "There has to be atleast one edge to save",
			Fields:  &[]string{"edges"},
		})
		return
	}

	// validate edges
	for idx, edge := range *jsonBody.Edges {
		if edge.Name == nil {
			WriteError(w, &AppError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Edge at %d does not have `name`", idx),
				Fields:  &[]string{fmt.Sprintf("edges.%d.name", idx)},
			})
			return
		}

		if edge.SrcId == 0 || edge.DestId == 0 {
			WriteError(w, &AppError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Edge at %d does not have `src_id` and `dest_id`", idx),
				Fields:  &[]string{fmt.Sprintf("edges.%d.name", idx)},
			})
			return
		}
	}

	saveErr := models.SaveMany(db, jsonBody.Edges)

	if saveErr != nil {
		WriteError(w, &AppError{
			Code:    http.StatusInternalServerError,
			Message: saveErr.Error(),
		})
		return
	}

	responseJson := make(map[string]string)

	responseJson["success"] = "true"

	WriteJson(w, responseJson, http.StatusOK)
}

func DeleteEdgesEndpoint(db *sqlx.DB, w http.ResponseWriter, r *http.Request) {

	var jsonBody EdgesListRequest

	if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	defer r.Body.Close()

	if len(*jsonBody.Edges) == 0 {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: "There has to be atleast one edge to save",
			Fields:  &[]string{"edges"},
		})
		return
	}

	// validate edges & add default values
	for idx, edge := range *jsonBody.Edges {

		if edge.Status == "" {
			edge.Status = models.ACTIVE
		}

		if edge.Name == nil {
			WriteError(w, &AppError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Edge at %d does not have `name`", idx),
				Fields:  &[]string{fmt.Sprintf("edges.%d.name", idx)},
			})
			return
		}

		if edge.SrcId == 0 || edge.DestId == 0 {
			WriteError(w, &AppError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Edge at %d does not have `src_id` and `dest_id`", idx),
				Fields:  &[]string{fmt.Sprintf("edges.%d.name", idx)},
			})
			return
		}
	}

	saveErr := models.DeleteMany(db, jsonBody.Edges)

	if saveErr != nil {
		WriteError(w, &AppError{
			Code:    http.StatusInternalServerError,
			Message: saveErr.Error(),
		})
		return
	}

	responseJson := make(map[string]string)

	responseJson["success"] = "true"

	WriteJson(w, responseJson, http.StatusOK)
}

type QueryRequest struct {
	Query string `json:"query"`
}

func RunQueryEndpoint(db *sqlx.DB, w http.ResponseWriter, r *http.Request) {

	var jsonBody QueryRequest

	if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	defer r.Body.Close()

	if jsonBody.Query == "" {
		WriteError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: "You must provide a query to execute",
			Fields:  &[]string{"query"},
		})
		return
	}

	edgeListPtr, queryErr := models.RunQuery(db, jsonBody.Query)

	if queryErr != nil {
		WriteError(w, &AppError{
			Code:    http.StatusInternalServerError,
			Message: queryErr.Error(),
		})
		return
	}

	responseJson := make(map[string]interface{})

	responseJson["success"] = "true"
	responseJson["edges"] = *edgeListPtr

	WriteJson(w, responseJson, http.StatusOK)
}
