package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)


type Data struct {
  From_location Location `json: "from_location"`
  To_location  Location  `json: "to_location"`
  Packages    Size  `json: "packages"`
}

type Size map[string]int
type Location map[string]string


type PriceSending struct {
  TariffID   int
  TariffName string
  Price      float64
  Delivery   int
  DaysMin    int
  DaysMax    int
}


func main() {
  from_location := "Россия, г. Москва, Cлавянский бульвар д.1"
  to_location := "Россия, Воронежская обл., г. Владивосток, ул. Ленина д.43"
  packages := Size{"height": 245, "weight": 100000, "length": 554, "width": 66}

  //whole := Data{from_location, to_location, object}
  token, err := auth()
  if err != nil {
    fmt.Print(err)
  }

  
  
  result, err := Calculate(from_location, to_location, packages, token)
  if err != nil {
    fmt.Println(err)
    return
  }

  for _, x := range result {

    fmt.Printf("Name: %s ||| Price: %.f ||| Days min/max: %d to %d|||\n", x.TariffName, x.Price, x.DaysMin, x.DaysMax)
  }
}

func Calculate(addrFrom string, addrTo string, packages Size, token string) ([]PriceSending, error) {
  // Формируем данные запроса
  obj := Data{
    From_location: Location{"address": addrFrom},
    To_location: Location {"address": addrTo},
    Packages: packages,
    }

toReq, err := json.Marshal(obj)
if err != nil {
  fmt.Printf("Happened this: %v", err)
  return nil, err
}
  
  // Формируем URL для запроса расчета стоимости
  url := "https://api.edu.cdek.ru/v2/calculator/tarifflist"

  // Выполняем запрос к API СДЭК и получаем ответ в виде []byte
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(toReq))
  if err != nil {
    fmt.Printf("POST REQUEST ERROR: %v", err)
    return nil, err
  }
 
  req.Header.Set("Authorization", "Bearer "+token)
  req.Header.Set("Content-Type", "application/json")
 
 
 

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Printf("error here: %v", err)
    return nil, err
  }

  // Парсим ответ от API СДЭК и возвращаем результат в виде структуры []PriceSending
  type Result struct {
    TariffCodes []struct {
      TariffCode       int     `json:"tariff_code"`
      TariffName       string  `json:"tariff_name"`
      TariffDescription string `json:"tariff_description"`
      DeliveryMode     int     `json:"delivery_mode"`
      DeliverySum      float64 `json:"delivery_sum"`
      PeriodMin        int     `json:"period_min"`
      PeriodMax        int     `json:"period_max"`
    } `json:"tariff_codes"`
  }

  var result Result
  if err := json.Unmarshal(body, &result); err != nil {
    fmt.Println("error here 151")
    return nil, err
  }

  prices := make([]PriceSending, 0, len(result.TariffCodes))
  for _, code := range result.TariffCodes {
    price := PriceSending{
      TariffID:   code.TariffCode,
      TariffName: code.TariffName,
      Price:      code.DeliverySum,
      DaysMin:    code.PeriodMin,
      DaysMax:    code.PeriodMax,
    }
    prices = append(prices, price)

  
  }
  return prices, nil
}



func auth() (accessToken string, err error) {
  account := "EMscd6r9JnFiQ3bLoyjJY6eM78JrJceI"
  password := "PjLZkKBHEiLK3YsjtNrt3TGNG0ahs3kG"
    apiURL := "https://api.edu.cdek.ru/v2/oauth/token?parameters"

    fmt.Println()
    values := url.Values{
        "grant_type":    {"client_credentials"},
        "client_id":     {account},
        "client_secret": {password},
    }

    resp, err := http.PostForm(apiURL, values)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()


  var authResp struct {
    AccessToken string `json:"access_token"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
    return "", err
  }
  

    return authResp.AccessToken, nil
}
