package send

import (
	"fmt"
	"github.com/go-chi/render"
	"infotecs-tz/internal/api/response"
	"log/slog"
	"net/http"
)

type Request struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type Response struct {
	response.Response
	TransactionHash string `json:"transactionHash"`
}

type Sender interface {
	Send(sender, recipient string, amount float64) (string, error)
	IsExists(address string) (bool, error)
}

func New(log *slog.Logger, sender Sender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Transactions.Send"

		log := log.With(slog.String("fn", fn))

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request body"))

			return
		}

		if req.From == "" || req.To == "" {
			log.Error("field 'from' or 'to' is empty")
			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, response.Error("field 'from' or 'to' is empty"))

			return
		}

		exists, err := sender.IsExists(req.From)
		if err != nil {
			log.Error(err.Error())
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		if exists == false {
			log.Error(fmt.Sprintf("%s is not exists", req.From))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error(fmt.Sprintf("%s is not exists", req.From)))

			return
		}

		exists, err = sender.IsExists(req.To)
		if err != nil {
			log.Error(err.Error())
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		if exists == false {
			log.Error(fmt.Sprintf("%s is not exists", req.To))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error(fmt.Sprintf("%s is not exists", req.To)))

			return
		}

		if req.Amount <= 0 {
			log.Error("amount less than 0")
			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, "amount must be greater than 0")

			return
		}

		txHash, err := sender.Send(req.From, req.To, req.Amount)
		if err != nil {
			log.Error(err.Error())
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		log.Info(fmt.Sprintf("success transaction %f from %s to %s", req.Amount, req.From, req.To))
		render.JSON(w, r, Response{
			Response:        response.OK(),
			TransactionHash: txHash,
		})
	}
}
