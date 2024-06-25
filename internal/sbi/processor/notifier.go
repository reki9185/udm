package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi/models"
	udm_context "github.com/free5gc/udm/internal/context"
	"github.com/free5gc/udm/internal/logger"
)

func (p *Processor) DataChangeNotificationProcedure(c *gin.Context,
	notifyItems []models.NotifyItem,
	supi string,
) {
	ctx, pd, err := udm_context.GetSelf().GetTokenCtx(models.ServiceName_NUDM_SDM, models.NfType_UDM)
	if err != nil {
		c.JSON(int(pd.Status), pd)
		return
	}

	ue, _ := udm_context.GetSelf().UdmUeFindBySupi(supi)

	clientAPI := p.consumer.GetSDMClient("DataChangeNotification")

	var problemDetails *models.ProblemDetails
	for _, subscriptionDataSubscription := range ue.UdmSubsToNotify {
		onDataChangeNotificationurl := subscriptionDataSubscription.OriginalCallbackReference
		dataChangeNotification := models.ModificationNotification{}
		dataChangeNotification.NotifyItems = notifyItems

		httpResponse, err := clientAPI.DataChangeNotificationCallbackDocumentApi.OnDataChangeNotification(
			ctx, onDataChangeNotificationurl, dataChangeNotification)
		if err != nil {
			if httpResponse == nil {
				logger.HttpLog.Error(err.Error())
				problemDetails = &models.ProblemDetails{
					Status: http.StatusForbidden,
					Detail: err.Error(),
				}
			} else {
				logger.HttpLog.Errorln(err.Error())

				problemDetails = &models.ProblemDetails{
					Status: int32(httpResponse.StatusCode),
					Detail: err.Error(),
				}
			}
		}
		defer func() {
			if rspCloseErr := httpResponse.Body.Close(); rspCloseErr != nil {
				logger.HttpLog.Errorf("OnDataChangeNotification response body cannot close: %+v", rspCloseErr)
			}
		}()
	}

	c.JSON(int(problemDetails.Status), problemDetails)
}

func (p *Processor) SendOnDeregistrationNotification(ueId string, onDeregistrationNotificationUrl string,
	deregistData models.DeregistrationData,
) *models.ProblemDetails {
	ctx, pd, err := udm_context.GetSelf().GetTokenCtx(models.ServiceName_NUDM_UECM, models.NfType_UDM)
	if err != nil {
		return pd
	}

	clientAPI := p.consumer.GetUECMClient("SendOnDeregistrationNotification")

	httpResponse, err := clientAPI.DeregistrationNotificationCallbackApi.DeregistrationNotify(
		ctx, onDeregistrationNotificationUrl, deregistData)
	if err != nil {
		if httpResponse == nil {
			logger.HttpLog.Error(err.Error())
			problemDetails := &models.ProblemDetails{
				Status: http.StatusInternalServerError,
				Cause:  "DEREGISTRATION_NOTIFICATION_ERROR",
				Detail: err.Error(),
			}

			return problemDetails
		} else {
			logger.HttpLog.Errorln(err.Error())
			problemDetails := &models.ProblemDetails{
				Status: int32(httpResponse.StatusCode),
				Cause:  "DEREGISTRATION_NOTIFICATION_ERROR",
				Detail: err.Error(),
			}

			return problemDetails
		}
	}
	defer func() {
		if rspCloseErr := httpResponse.Body.Close(); rspCloseErr != nil {
			logger.HttpLog.Errorf("DeregistrationNotify response body cannot close: %+v", rspCloseErr)
		}
	}()

	return nil
}
