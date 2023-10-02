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

type Professores struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Registro      int                `json:"registro" bson:"registro"`
	Nome          string             `json:"nome" bson:"nome"`
	Sobrenome     string             `json:"sobrenome" bson:"sobrenome"`
	Telefone      int                `json:"telefone" bson:"telefone"`
}

type ProfessoresHandler struct {
	Col dbiface.Collection
}

func inserirProfessor(ctx context.Context, professores []Professores, collection dbiface.Collection) ([]interface{}, *echo.HTTPError) {
	var insertedIDs []interface{}

	for _, professor := range professores {
		professor.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, professor)
		if err != nil {
			log.Errorf("Unable to insert: %v", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to connect to databse")
		}
		insertedIDs = append(insertedIDs, insertID.InsertedID)
	}
	return insertedIDs, nil
}

func (uh *ProfessoresHandler) InserirProfessor(c echo.Context) error {
	var professores []Professores

	if err := c.Bind(&professores); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return c.JSON(http.StatusBadRequest, "Unable to parse request payload")
	}

	IDs, err := inserirProfessor(context.Background(), professores, uh.Col)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, IDs)
}

func buscarProfessores(ctx context.Context, q url.Values, collection dbiface.Collection) ([]Professores, *echo.HTTPError) {
	var professores []Professores

	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	if filter["_id"] != nil {
		docID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			return professores, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
		}
		filter["_id"] = docID
	}
	cursor, err := collection.Find(ctx, bson.M(filter)) /*filter, que corresponde ao q url.Values, está sendo
	convertido para bson. Filter é argumento do método Find e está sendo definido logo acima*/
	if err != nil {
		log.Errorf("Unable to find the teacher: %v", err)
		return professores, echo.NewHTTPError(http.StatusNotFound, "Unable to find the teacher")
	}
	err = cursor.All(ctx, &professores)
	if err != nil {
		log.Errorf("Unable to read the cursor: %v", err)
		return professores, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	}
	return professores, nil
}

func (uh *ProfessoresHandler) BuscarProfessores(c echo.Context) error {
	professores, err := buscarProfessores(context.Background(), c.QueryParams(), uh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, professores)
}

func buscarProfessor(ctx context.Context, id string, collection dbiface.Collection) (Professores, *echo.HTTPError) {
	var professores Professores
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return professores, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert do ObjectID")
	}
	res := collection.FindOne(ctx, bson.M{"_id": docID})
	err = res.Decode(&professores)
	if err != nil {
		return professores, echo.NewHTTPError(http.StatusNotFound, "Unable to find the teacher")
	}
	return professores, nil
}

func (uh *ProfessoresHandler) BuscarProfessor(c echo.Context) error {
	professores, err := buscarProfessor(context.Background(), c.Param("id"), uh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, professores)
}

func alterarProfessor(ctx context.Context, id string, reqBody io.ReadCloser, collection dbiface.Collection) (Professores, *echo.HTTPError) {
	var professores Professores

	//convertendo o id, que é um string, para primitive.ObjectID
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Cannot convert to ObjectID: %v", err)
		return professores, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}

	//procurar se o professor existe, caso contrário erro 404 (Not Found)
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&professores); err != nil {
		log.Errorf("Unable to decode to teacher: %v", err)
		return professores, echo.NewHTTPError(http.StatusNotFound, "Unable to find the teacher")
	} /*todos os valores, obtidos do método FindOne e atribuídos para res, são decodificados para professores,
	ou seja, professores está sendo populado*/

	//JSON decodificação do reqBody
	if err := json.NewDecoder(reqBody).Decode(&professores); err != nil {
		log.Errorf("Unable to decode using reqBody: %v", err)
		return professores, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	} /*ler o reqBody e usar as informações inseridas para popular os campos de professores, que "representa" a
	struct Professores*/

	//atualização do professor
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": professores}) /* _, err, pois UpdateOne possui 2
	retornos, *mongo.UpdateResult e error, sendo que não necessita-se aqui do UpdateResult*/
	if err != nil {
		log.Errorf("Unable to update the teacher: %v", err)
		return professores, echo.NewHTTPError(http.StatusInternalServerError, "Unable to update the teacher")
	}
	return professores, nil
}

func (uh *ProfessoresHandler) AtualizarProfessor(c echo.Context) error {
	professores, err := alterarProfessor(context.Background(), c.Param("id"), c.Request().Body, uh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, professores)
}

func deletarProfessor(ctx context.Context, id string, collection dbiface.Collection) (int64, *echo.HTTPError) {
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to delete the teacher: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	res, err := collection.DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, "Unable to delete the teacher")
	}
	return res.DeletedCount, nil
}

func (uh *ProfessoresHandler) DeletarProfessor(c echo.Context) error {
	del, err := deletarProfessor(context.Background(), c.Param("id"), uh.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, del)
}
