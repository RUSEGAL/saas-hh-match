package tokens

import (
	dbtokens "go-api/internal/repository/tokens"
)

func GenerateToken(username string, telegramID int64) (string, error) {

	return dbtokens.AddTokensDb(username, telegramID)
}

func FindToken(token string) (int, error) {

	return dbtokens.FindTokenDb(token)
}

func RegenerateToken() {

}

func DeleteToken() {

}
