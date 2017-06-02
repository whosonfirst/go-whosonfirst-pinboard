package whosonfirst

type Config struct {
     DataRoot string
     AuthToken string
}

func NewDefaultConfig () (*Config, error) {

     c := Config{
     	DataRoot: "https://whosonfirst.mapzen.com/data",
	AuthToken: "",
     }

     return &c, nil
}