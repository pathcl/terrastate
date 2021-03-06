package handler

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dchest/safefile"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/webhippie/terrastate/pkg/config"
	"github.com/webhippie/terrastate/pkg/helper"
)

// Update is used to update a specific state.
func Update(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer handleMetrics(time.Now(), "update", chi.URLParam(req, "*"))

		dir := strings.Replace(
			path.Join(
				cfg.Server.Storage,
				chi.URLParam(req, "*"),
			),
			"../", "", -1,
		)

		full := path.Join(
			dir,
			"terraform.tfstate",
		)

		content, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Info().
				Err(err).
				Msg("failed to load request body")

			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)

			return
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Info().
				Err(err).
				Str("dir", dir).
				Msg("failed to create state dir")

			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)

			return
		}

		if cfg.General.Secret != "" {
			encrypted, err := helper.Encrypt(content, []byte(cfg.General.Secret))

			if err != nil {
				log.Info().
					Err(err).
					Str("file", full).
					Msg("failed to encrypt the state")

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)

				return
			}

			content = encrypted
		}

		if _, err := os.Stat(full); os.IsNotExist(err) {
			if err := safefile.WriteFile(full, content, 0644); err != nil {
				log.Info().
					Err(err).
					Str("file", full).
					Msg("failed to create state file")

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)

				return
			}

			log.Info().
				Str("file", full).
				Msg("successfully created state file")
		} else {
			if err := safefile.WriteFile(full, content, 0644); err != nil {
				log.Info().
					Err(err).
					Str("file", full).
					Msg("failed to update state file")

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)

				return
			}

			log.Info().
				Str("file", full).
				Msg("successfully updated state file")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
