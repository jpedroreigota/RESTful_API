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

type Disciplinas struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Nome         string             `json:"nome" bson:"nome"`
	CargaHoraria int                `json:"cargaHoraria" bson:"cargaHoraria"`
}

type DisciplinasHandler struct {
	Col dbiface.Collection
}

func inserirDisciplina(ctx context.Context, disciplinas []Disciplinas, collection dbiface.Collection) ([]interface{}, *echo.HTTPError) {
	var insertedIDs []interface{}

	for _, disciplina := range disciplinas {
		disciplina.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, disciplina)
		if err != nil {
			log.Errorf("Unable to insert: %v", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to connect to database")
		}
		insertedIDs = append(insertedIDs, insertID.InsertedID)
	}
	return insertedIDs, nil
}

func (oh *DisciplinasHandler) InserirDisciplina(c echo.Context) error {
	var disciplinas []Disciplinas

	if err := c.Bind(&disciplinas); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return c.JSON(http.StatusBadRequest, "Unable to parse request payload")
	}
	IDs, err := inserirDisciplina(context.Background(), disciplinas, oh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, IDs)
}

func buscarDisciplinas(ctx context.Context, q url.Values, collection dbiface.Collection) ([]Disciplinas, *echo.HTTPError) {
	var disciplinas []Disciplinas
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	if filter["_id"] != nil { //convertendo o id, que no filter é um string, para um primitiveObjectID
		docID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			return disciplinas, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
		}
		filter["_id"] = docID
	}
	cursor, err := collection.Find(ctx, bson.M(filter)) /*filter, que corresponde ao q url.Values, está sendo
	convertido para bson. Filter é argumento do método Find e está sendo definido logo acima*/
	if err != nil {
		log.Errorf("Unable to find the discipline: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusNotFound, "Unable to find the discipline")
	}
	err = cursor.All(ctx, &disciplinas)
	if err != nil {
		log.Errorf("Unable to read the cursor: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	}
	return disciplinas, nil
}

func (oh *DisciplinasHandler) BuscarDisciplinas(c echo.Context) error {
	disciplinas, err := buscarDisciplinas(context.Background(), c.QueryParams(), oh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, disciplinas)
}

func buscarDisciplina(ctx context.Context, id string, collection dbiface.Collection) (Disciplinas, *echo.HTTPError) {
	var disciplinas Disciplinas
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return disciplinas, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)
	err = res.Decode(&disciplinas)
	if err != nil {
		return disciplinas, echo.NewHTTPError(http.StatusNotFound, "Unable to find the discipline")
	}
	return disciplinas, nil
}

func (oh *DisciplinasHandler) BuscarDisciplina(c echo.Context) error {
	disciplinas, err := buscarDisciplina(context.Background(), c.Param("id"), oh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, disciplinas)
}

func atualizarDisciplina(ctx context.Context, id string, reqBody io.ReadCloser, collection dbiface.Collection) (Disciplinas, *echo.HTTPError) {
	var disciplinas Disciplinas

	//convertendo o id, que é um string, para primitive.ObjectID
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Cannot convert to ObjectID: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}

	//procurar se o disciplina existe, caso contrário erro 404 (Not Found)
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&disciplinas); err != nil {
		log.Errorf("Unable to decode to discipine: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusNotFound, "Unable to find the discipline")
	} /*todos os valores, obtidos do método FindOne e atribuídos para res, são decodificados para disciplinas,
	ou seja, disciplinas está sendo populado*/

	//JSON decodificação do reqBody
	if err := json.NewDecoder(reqBody).Decode(&disciplinas); err != nil {
		log.Errorf("Unable to decode using reqBody: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	} /*ler o reqBody e usar as informações inseridas para popular os campos de disciplinas, que "representa" a
	struct disciplinas*/

	//validação da requisição
	if err := v.Struct(disciplinas); err != nil {
		log.Errorf("Unable to validate the struct: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusBadRequest, "Unable to validate request payload")
	} //verificar se o produto, agora atualizado, é válido ou não

	//atualização do disciplina
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": disciplinas}) /* _, err, pois UpdateOne possui 2
	retornos, *mongo.UpdateResult e error, sendo que não necessita-se aqui do UpdateResult*/
	if err != nil {
		log.Errorf("Unable to update the discipline: %v", err)
		return disciplinas, echo.NewHTTPError(http.StatusInternalServerError, "Unable to update the discipline")
	}
	return disciplinas, nil
}

func (oh *DisciplinasHandler) AtualizarDisciplina(c echo.Context) error {
	disciplinas, err := atualizarDisciplina(context.Background(), c.Param("id"), c.Request().Body, oh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, disciplinas)
}

func deletarDisciplina(ctx context.Context, id string, collection dbiface.Collection) (int64, *echo.HTTPError) {
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to delete the discipline: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	filter := bson.M{"_id": docID}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, "Unable to delete the discipline")
	}
	return res.DeletedCount, nil
}

func (oh *DisciplinasHandler) DeletarDisciplina(c echo.Context) error {
	del, err := deletarDisciplina(context.Background(), c.Param("id"), oh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, del)
}
