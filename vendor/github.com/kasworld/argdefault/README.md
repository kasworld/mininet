# make commandline argument from struct field and tag and set field value 

struct 의 field와 tag로 부터 명령행 인자의 이름과 default값을 자동으로 만들어 냅니다. 

또 명령행 인자로 주어진 값을 자동으로 struct의 해당 field에 넣어 줍니다. 

# 기능 

golang reflect를 사용해서 struct 의 field 이름과 tag 내용으로 

commandline에서 읽을 인자의 이름과 default 값을 자동으로 생성해주고 

또 인자로 받은 값을 struct의 filed에 자동으로 넣어 줍니다. 

이를 사용해서 프로그램의 config struct 가 변경될때마다 

해당하는 필드를 명령행에서 읽어 들이는 코드를 변경하는 수고를 덜 수 있습니다. 

## struct filed의 tag 내용 

default : 인자가 없는 경우 사용할 기본 값 

argname : 이 태그가 있는 경우 이 필드가 commandline argument로 부터 설정됩니다. 

    빈 값인 경우 : 필드 이름이 argument 이름이 됩니다. 
    있는 경우 : 이 값이 argument 이름이 됩니다. 

## 지원 하는 필드 타입 

    int, int8, int16, int32, int64
    uint, uint8, uint16, uint32, uint64
    float64, float32
    bool
    string


# 사용 예제 

example/main.go 참고 

    struct가  아래와 같이 정의된 경우 
    type Config struct {
        BaseLogDir            string  `argname:""`
        DataFolder            string  `default:"./serverdata"`
        ClientDataFolder      string  `default:"./clientdata" argname:""`
        GroundRPC             string  `default:"localhost:14002"  argname:""`
        ServicePort           string  `default:":14101"  argname:"port"`
        AdminPort             string  `default:":14201"  argname:""`
        TowerFilename         string  `default:"starting" argname:""`
        TowerNumber           int     `default:"1" argname:""`
        DisplayName           string  `default:"Default" argname:""`
        ConcurrentConnections int     `default:"10000" argname:""`
        ActTurnPerSec         float64 `default:"2.0" argname:""`
        StandAlone            bool    `default:"true" argname:""`
    }

    자동으로 아래와 같이 명령행 인자를 정의 해줍니다. 
    또 해당 인자가 명령행으로 주어진 경우 그 값을 struct의 해당 field에 넣어 줍니다. 
    -ActTurnPerSec float
            ActTurnPerSec (default 2)
    -AdminPort string
            AdminPort (default ":14201")
    -BaseLogDir string
            BaseLogDir
    -ClientDataFolder string
            ClientDataFolder (default "./clientdata")
    -ConcurrentConnections int
            ConcurrentConnections (default 10000)
    -DisplayName string
            DisplayName (default "Default")
    -GroundRPC string
            GroundRPC (default "localhost:14002")
    -StandAlone
            StandAlone (default true)
    -TowerFilename string
            TowerFilename (default "starting")
    -TowerNumber int
            TowerNumber (default 1)
    -port string
            ServicePort (default ":14101")
