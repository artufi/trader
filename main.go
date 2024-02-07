package main

type Config struct {
	XTB struct {
		AuthURL string `yaml:"authURL"`
	} `yaml:"XTB"`
}

func main() {
	//plik z logami
	//logger inicjalizacja
	//wczytanie konfiguracji
	//uruchomienie serwera HTTP nasluchujacego na requesty z innego serwisu
}
