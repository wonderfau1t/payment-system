package get_balance

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"infotecs-tz/internal/api/response"
	"infotecs-tz/internal/storage"

	"log/slog"
	"net/http"
)

type Response struct {
	response.Response
	Balance float64 `json:"balance"`
}

type BalanceGetter interface {
	GetBalance(address string) (float64, error)
}

func New(log *slog.Logger, getter BalanceGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Transactions.GetBalance"
		log := log.With(slog.String("fn", fn))

		address := chi.URLParam(r, "address")
		if address == "" {
			log.Info("address is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("address is empty"))

			return
		}

		balance, err := getter.GetBalance(address)
		if err != nil {
			if errors.Is(err, storage.ErrAddressNotFound) {
				log.Info(err.Error())
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.Error(err.Error()))

				return
			}
			log.Error(err.Error())
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
		}

		log.Info(fmt.Sprintf("%s: %f", address, balance))
		render.JSON(w, r, Response{
			Response: response.OK(),
			Balance:  balance,
		})

		return
	}
}
