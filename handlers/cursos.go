package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/krunal4amity/tronicscorp/dbiface"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Cursos struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty`
	Nome string             `json:"nome" bson:"nome"`
}

type CursosHandler struct {
	Col dbiface.Collection
}

func inserirCurso(ctx context.Context, cursos []Cursos, collection dbiface.Collection) ([]interface{}, *echo.HTTPError) {
	var insertedIds []interface{}
	for _, curso := range cursos {
		curso.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, curso)
		if err != nil {
			log.Errorf("Unable to insert: %v", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to connect to database")
		}
		insertedIds = append(insertedIds, insertID.InsertedID)
	}
	return insertedIds, nil
}

func (ah *CursosHandler) InserirCurso(c echo.Context) error {
	var cursos []Cursos

	if err := c.Bind(&cursos); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to parse the request payload")
	}

	IDs, err := inserirCurso(context.Background(), cursos, ah.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, IDs)
}

func buscarCursos(ctx context.Context, q url.Values, collection dbiface.Collection) ([]Cursos, *echo.HTTPError) {
	var cursos []Cursos
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	if filter["_id"] != nil { //convertendo o id, que no filter é um string, para um primitiveObjectID
		docID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			return cursos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
		}
		filter["_id"] = docID
	}
	cursor, err := collection.Find(ctx, bson.M(filter)) /*filter, que corresponde ao q url.Values, está sendo
	convertido para bson. Filter é argumento do método Find e está sendo definido logo acima*/
	if err != nil {
		log.Errorf("Unable to find the course: %v", err)
		return cursos, echo.NewHTTPError(http.StatusNotFound, "Unable to find the course")
	}
	err = cursor.All(ctx, &cursos)
	if err != nil {
		log.Errorf("Unable to read the cursor: %v", err)
		return cursos, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	}
	return cursos, nil
}

func (ah *CursosHandler) BuscarCursos(c echo.Context) error {
	cursos, err := buscarCursos(context.Background(), c.QueryParams(), ah.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, cursos)
}

func buscarCurso(ctx context.Context, id string, collection dbiface.Collection) (Cursos, *echo.HTTPError) {
	var cursos Cursos
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return cursos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	filter := bson.M{"id": docID}
	res := collection.FindOne(ctx, filter)
	err = res.Decode(&cursos)
	if err != nil {
		return cursos, echo.NewHTTPError(http.StatusNotFound, "Unable to find the course")
	}
	return cursos, nil
}

func (h *CursosHandler) BuscarCurso(c echo.Context) error {
	cursos, err := buscarCurso(context.Background(), c.Param("id"), h.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, cursos)
}

func atualizarCurso(ctx context.Context, id string, reqBody io.ReadCloser, collection dbiface.Collection) (Cursos, *echo.HTTPError) {
	var cursos Cursos

	//convertendo o id, que é um string, para primitive.ObjectID
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Cannot convert to ObjectID: %v", err)
		return cursos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}

	//procurar se o curso existe, caso contrário erro 404 (Not Found)
	filter := bson.M{"id": docID}
	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&cursos); err != nil {
		log.Errorf("Unable to decode to course: %v", err)
		return cursos, echo.NewHTTPError(http.StatusNotFound, "Unable to find the course")
	} /*todos os valores, obtidos do método FindOne e atribuídos para res, são decodificados para cursos,
	ou seja, cursos está sendo populado*/

	//JSON decodificação do reqBody
	if err := json.NewDecoder(reqBody).Decode(&cursos); err != nil {
		log.Errorf("Unable to decode using reqBody: %v", err)
		return cursos, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	} /*ler o reqBody e usar as informações inseridas para popular os campos de cursos, que "representa" a
	struct cursos*/

	//validação da requisição
	if err := v.Struct(cursos); err != nil {
		log.Errorf("Unable to validate the struct: %v", err)
		return cursos, echo.NewHTTPError(http.StatusBadRequest, "Unable to validate request payload")
	} //verificar se o produto, agora atualizado, é válido ou não

	//atualização do curso
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": cursos}) /* _, err, pois UpdateOne possui 2
	retornos, *mongo.UpdateResult e error, sendo que não necessita-se aqui do UpdateResult*/
	if err != nil {
		log.Errorf("Unable to update the course: %v", err)
		return cursos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to update the course")
	}
	return cursos, nil
}

func (ah *CursosHandler) AtualizarCurso(c echo.Context) error {
	cursos, err := atualizarCurso(context.Background(), c.Param("id"), c.Request().Body, ah.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, cursos)
}

func deletarCurso(ctx context.Context, id string, collection dbiface.Collection) (int64, *echo.HTTPError) {
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to delete the course: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	res, err := collection.DeleteOne(ctx, bson.M{"id": docID})
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, "Unable to delete the course")
	}
	return res.DeletedCount, nil
}

func (h *CursosHandler) DeletarCurso(c echo.Context) error {
	delCount, err := deletarCurso(context.Background(), c.Param("id"), h.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, delCount)
}
