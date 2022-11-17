package routes

import (
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen/kamailioaddress"
	"net/http"
	"strconv"
)

// @Description used to identify a carrier by the IP address of incomming traffic
type GetCarrierAddress struct {
	// the id for this specific IP record
	ID int `json:"ID"`
	// IP address of the carrier
	IpAddr string `json:"ip_address" example:"192.168.0.1"`
	// number of bits to mask the ip address, used to match entire ip subnets
	Mask int8 `json:"mask" example:"32" format:"int8"`
	// port of the expected traffic comes from, 0 means any port
	Port int16 `json:"port" example:"5060" format:"int16"`
	// name of the carrier this ip is to be identified as
	Carrier string `json:"carrier_name"  example:"twilio"`
}

// @Description Identical to GetCarrierAddress except the ID is omitted
type AddCarrierAddress struct {
	// IP address of the carrier
	IpAddr string `json:"ip_address" example:"192.168.0.1"`
	// number of bits to mask the ip address, used to match entire ip subnets
	Mask int8 `json:"mask" example:"32" format:"int8"`
	// port of the expected traffic comes from, 0 means any port
	Port int16 `json:"port" example:"5060" format:"int16"`
	// name of the carrier this ip is to be identified as
	Carrier string `json:"carrier_name"  example:"twilio"`
}

type carrierAddressRoute struct {
	conf   *c.Config
	client *gen.Client
}

// @Router      /carrier_address/{id} [get]
// @Description gets ip address record
// @Tags		carrier_address
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the ip address entry"
// @Success     200      {object}	GetCarrierAddress
// @Failure     400		 {object}	HTTPError
func (car *carrierAddressRoute) getAddress(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	carrierAddress, err := car.client.KamailioAddress.Get(c, cid)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetCarrierAddress
	response.ID = carrierAddress.ID
	response.IpAddr = carrierAddress.IPAddr
	response.Mask = carrierAddress.Mask
	response.Port = carrierAddress.Port
	response.Carrier = carrierAddress.Tag

	c.JSON(http.StatusOK, &response)
}

// @Router		/carrier_address/{id} [delete]
// @Description deletes ip address record
// @Tags		carrier_address
// @Accept      json
// @Produce     json
// @Param       id       path     int true "ID of the ip address entry to delete"
// @Success     200      {object}	HTTPError
// @Failure     400		 {object}	HTTPError
func (car *carrierAddressRoute) deleteAddress(c *gin.Context) {
	id := c.Param("id")
	cid, err := strconv.Atoi(id)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	err = car.client.KamailioAddress.DeleteOneID(cid).Exec(c)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, &HTTPError{http.StatusOK, "Deletion Successful"})
}

// @Router      /carrier_address [get]
// @Description gets all the ip address for carrier
// @Tags		carrier_address
// @Accept      json
// @Produce     json
// @Param       carrier_name  query     string false "carrier to filter ip result set against" example(twilio)
// @Success     200      {array}	GetCarrierAddress
// @Failure     400		 {object}	HTTPError
func (car *carrierAddressRoute) getAddressList(c *gin.Context) {
	carrierName := c.Query("carrier_name")

	query := car.client.KamailioAddress.Query()

	if carrierName != "" {
		query = query.Where(kamailioaddress.TagEQ(carrierName))
	}

	carrierAddressList, err := query.All(c)

	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var resultList = make([]*GetCarrierAddress, len(carrierAddressList))
	for i := 0; i < len(carrierAddressList); i++ {
		resultList[i] = &GetCarrierAddress{
			ID:      carrierAddressList[i].ID,
			IpAddr:  carrierAddressList[i].IPAddr,
			Mask:    carrierAddressList[i].Mask,
			Port:    carrierAddressList[i].Port,
			Carrier: carrierAddressList[i].Tag,
		}
	}
	c.JSON(http.StatusOK, resultList)
}

// @Router      /carrier_address [post]
// @Description gets ip address for carrier
// @Tags		carrier_address
// @Accept      json
// @Produce     json
// @Param       message  body  AddCarrierAddress false "carrier adddress record to insert into the database"
// @Success     200      {object}	GetCarrierAddress
// @Failure     400		 {object}	HTTPError
func (car *carrierAddressRoute) addAddress(c *gin.Context) {
	var newCarrierAddress AddCarrierAddress
	if err := c.ShouldBindJSON(&newCarrierAddress); err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	carrierAddress, err := car.client.KamailioAddress.Create().
		SetIPAddr(newCarrierAddress.IpAddr).
		SetMask(newCarrierAddress.Mask).
		SetPort(newCarrierAddress.Port).
		SetTag(newCarrierAddress.Carrier).
		Save(c)
	if err != nil {
		NewError(c, http.StatusBadRequest, err)
		return
	}

	var response GetCarrierAddress
	response.ID = carrierAddress.ID
	response.IpAddr = carrierAddress.IPAddr
	response.Mask = carrierAddress.Mask
	response.Port = carrierAddress.Port
	response.Carrier = carrierAddress.Tag

	c.JSON(http.StatusOK, &response)
}

func addAddressRoutes(conf *c.Config, client *gen.Client, rg *gin.RouterGroup) {

	car := new(carrierAddressRoute)
	car.client = client
	car.conf = conf

	rg.GET("carrier_address/:id", car.getAddress)
	rg.GET("carrier_address", car.getAddressList)
	rg.POST("carrier_address", car.addAddress)
	rg.DELETE("carrier_address/:id", car.deleteAddress)

}
