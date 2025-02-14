package get_last

import (
	"github.com/go-chi/render"
	"infotecs-tz/internal/api/response"
	"infotecs-tz/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
)

type Response struct {
	response.Response
	TransactionsCount int                   `json:"transactionsCount"`
	Transactions      []storage.Transaction `json:"transactions"`
}

type Storage interface {
	GetLast(limit int) ([]storage.Transaction, error)
}

func New(log *slog.Logger, storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Transactions.GetLast"
		log := log.With(slog.String("fn", fn))

		countStr := r.URL.Query().Get("count")
		if countStr == "" {
			log.Error("count is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("count is empty"))

			return
		}

		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			log.Error(err.Error())
			render.Status(r, http.StatusUnprocessableEntity)
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		transactions, err := storage.GetLast(count)
		if err != nil {
			log.Error(err.Error())
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		render.JSON(w, r, Response{
			Response:          response.OK(),
			TransactionsCount: len(transactions),
			Transactions:      transactions,
		})
	}
}
