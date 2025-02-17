package endpoints

import (
  "net/http"
  "encoding/json"
  "strings"
  "golang.org/x/crypto/bcrypt"
  "postman-twitter/util"
  "postman-twitter/models"
  "postman-twitter/auth"
  "postman-twitter/redis"
)

func SignUpHandler(r *http.Request) (interface{}, *util.HTTPError) {
  var userAuth models.UserAuth
  err := json.NewDecoder(r.Body).Decode(&userAuth)
  if err != nil {
      return nil, util.BadRequest(util.BAD_JSON_ERROR)
  }
  if userAuth.Username == "" || userAuth.Password == "" {
    return nil, util.BadRequest(util.DECODING_ERROR)
  }

  //Salting and Hashing Password
  var hashedPassword []byte
  hashedPassword, err = bcrypt.GenerateFromPassword([]byte(userAuth.Password), 8)
  if err != nil {
    return nil, util.InternalServerError(util.SALTING_ERROR)
  }
  userAuth.Password = string(hashedPassword)
  err = models.StoreAuthDetails(userAuth)
  if err != nil {
    return nil, util.InternalServerError(util.USER_ALREADY_EXIST_ERROR)
  }
  return util.GENERIC_SUCCESS_RESPONSE, nil
}

func SignInHandler(r *http.Request) (interface{}, *util.HTTPError) {
  var userAuth models.UserAuth
  err := json.NewDecoder(r.Body).Decode(&userAuth)
  if err != nil {
      return nil, util.BadRequest(util.BAD_JSON_ERROR)
  }
  if userAuth.Username == "" || userAuth.Password == "" {
    return nil, util.BadRequest(util.DECODING_ERROR)
  }

  var existingUserAuth models.UserAuth
  existingUserAuth, err = models.GetAuthDetails(userAuth.Username)
  if err != nil {
    return nil, util.Unauthorized(util.USER_DOES_NOT_EXIST_ERROR)
  }

  if err = bcrypt.CompareHashAndPassword([]byte(existingUserAuth.Password), []byte(userAuth.Password)); err != nil {
    return nil, util.Unauthorized(util.MISMATCH_PASSWORD_ERROR)
	}

  var jwtToken string
  jwtToken, err = auth.CreateJWTAuth(existingUserAuth)
  if err != nil {
    return nil, util.InternalServerError(util.JWT_ERROR)
  }
  payload := map[string]interface{}{
    "jwt_token": jwtToken,
    "user_id": existingUserAuth.ID,
  }
  return payload, nil
}

func LogOutHandler(r *http.Request) (interface{}, *util.HTTPError) {
	authStrings := strings.SplitN(r.Header.Get("Authorization"), " ", 3)
  bearerToken := authStrings[1]
  err := redis.StoreKeyValue(bearerToken, bearerToken)
  if err != nil {
    return nil, util.InternalServerError(util.LOGOUT_SESSION_ERROR)
  }
  return util.GENERIC_SUCCESS_RESPONSE, nil
}
