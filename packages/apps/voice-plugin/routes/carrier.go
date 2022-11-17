package routes

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"net/http"
	"strconv"
)

// @Description: Information required to make outbound calls with this carrier
type GetCarrier struct {
	// the id for this specific carrier record
	ID int `json:"ID"`
	// username to use when authenticating with the carrier
	Username string `json:"username"  example:"my-byoc-twilio-user"`
	// ream the carrier uses in its digest challenge
	Realm string `json:"realm"  example:"sip.twilio.com"`
	// ream the carrier uses in its digest challenge
	Domain string `json:"domain"  example:"my-byoc-account.pstn.twilio.com"`
	// name of the carrier this number is associated with
	Carrier string `json:"carrier_name"  example:"twilio"`
}

// @Description: Information required to make outbound calls with this carrier
type AddCarrier struct {
	// username to use when authenticating with the carrier
	Username string `json:"username"  example:"my-byoc-twilio-user"`
	// password to use when authenticating with the carrier
	Password string `json:"password"  example:"my-secret-password"`
	// ream the carrier uses in its digest challenge
	Realm string `json:"realm"  example:"sip.twilio.com"`
	// ream the carrier uses in its digest challenge
	Domain string `json:"domain"  example:"my-byoc-account.pstn.twilio.com"`
	// name of the carrier this number is associated with
	Carrier string `json:"carrier_name"  example:"twilio"`
}

type carrierRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /carrier/{id} [get]
// @Description gets a carrier record
// @Tags		carrier
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the carrier record"
// @Success     200      {object}	GetCarrier
// @Failure     400		 {object}	HTTPError
func (cr *carrierRoute) getCarrier(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	carrier, err := cr.client.OpenlineCarrier.Get(c, cid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetCarrier
	response.ID = carrier.ID
	response.Domain = carrier.Domain
	response.Realm = carrier.Realm
	response.Username = carrier.Username
	response.Carrier = carrier.CarrierName

	c.JSON(http.StatusOK, response)
}

// @Router      /carrier/{id} [delete]
// @Description deletes a carrier record
// @Tags		carrier
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the carrier to delete"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (cr *carrierRoute) deleteCarrier(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	err = cr.client.OpenlineCarrier.DeleteOneID(cid).Exec(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

// @Router      /carrier [get]
// @Description gets a list carrier records
// @Tags		carrier
// @Accept      json
// @Produce     json
// @Success     200      {array}	GetCarrier
// @Failure     400		 {object}	HTTPError
func (cr *carrierRoute) getCarrierList(c *gin.Context) {

	carrierList, err := cr.client.OpenlineCarrier.Query().All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	resultList := make([]*GetCarrier, len(carrierList))

	for i := 0; i < len(carrierList); i++ {
		resultList[i] = &GetCarrier{
			ID:       carrierList[i].ID,
			Domain:   carrierList[i].Domain,
			Realm:    carrierList[i].Realm,
			Username: carrierList[i].Username,
			Carrier:  carrierList[i].CarrierName,
		}
	}

	c.JSON(http.StatusOK, &resultList)
}

// @Router      /carrier [post]
// @Description creates a new carrier
// @Tags		carrier
// @Accept      json
// @Produce     json
// @Param       message  body  AddCarrier false "carrier to insert into the database"
// @Success     200      {object}	GetCarrier
// @Failure     400		 {object}	HTTPError
func (cr *carrierRoute) addCarrier(c *gin.Context) {
	var newCarrier AddCarrier
	if err := c.ShouldBindJSON(&newCarrier); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}
	data := md5.Sum([]byte(newCarrier.Username + ":" + newCarrier.Realm + ":" + newCarrier.Password))
	ha1 := fmt.Sprintf("%x", data)
	carrier, err := cr.client.OpenlineCarrier.Create().
		SetCarrierName(newCarrier.Carrier).
		SetUsername(newCarrier.Username).
		SetDomain(newCarrier.Domain).
		SetHa1(ha1).
		SetRealm(newCarrier.Realm).
		Save(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetCarrier
	response.ID = carrier.ID
	response.Domain = carrier.Domain
	response.Realm = carrier.Realm
	response.Username = carrier.Username
	response.Carrier = carrier.CarrierName

	c.JSON(http.StatusOK, response)
}

func addCarrierRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {

	cr := new(carrierRoute)
	cr.conf = conf
	cr.client = client

	rg.GET("carrier/:id", cr.getCarrier)
	rg.GET("carrier", cr.getCarrierList)
	rg.POST("carrier", cr.addCarrier)
	rg.DELETE("carrier/:id", cr.deleteCarrier)

}
