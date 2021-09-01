module gitlab.com/verygoodsoftwarenotvirus/todo

go 1.16

require (
	cloud.google.com/go/storage v1.15.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3
	github.com/Azure/azure-storage-blob-go v0.13.0
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/GuiaBolso/darwin v0.0.0-20191218124601-fd6d2aa3d244
	github.com/Masterminds/squirrel v1.5.0
	github.com/PuerkitoBio/goquery v1.6.1 // indirect
	github.com/RoaringBitmap/roaring v0.6.0 // indirect
	github.com/aead/chacha20poly1305 v0.0.0-20201124145622-1a5aba2a8b29 // indirect
	github.com/alexedwards/argon2id v0.0.0-20210326052512-e2135f7c9c77
	github.com/alexedwards/scs/mysqlstore v0.0.0-20210407073823-f445396108a4
	github.com/alexedwards/scs/postgresstore v0.0.0-20210407073823-f445396108a4
	github.com/alexedwards/scs/sqlite3store v0.0.0-20210407073823-f445396108a4
	github.com/alexedwards/scs/v2 v2.4.0
	github.com/andybalholm/cascadia v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.38.35
	github.com/blevesearch/bleve/v2 v2.0.3
	github.com/boombuler/barcode v1.0.1
	github.com/brianvoe/gofakeit/v5 v5.11.2
	github.com/carolynvs/magex v0.5.0
	github.com/cznic/ql v1.2.0 // indirect
	github.com/emicklei/hazana v1.9.6
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/cors v1.2.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.2.0
	github.com/google/wire v0.5.0
	github.com/gorilla/securecookie v1.1.1
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/lib/pq v1.10.1
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/magefile/mage v1.11.0
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/makiuchi-d/gozxing v0.0.0-20210324052758-57132e828831
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/mikespook/gorbac v2.1.0+incompatible // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moul/http2curl v1.0.0
	github.com/mssola/user_agent v0.5.2
	github.com/mxschmitt/playwright-go v0.1100.0
	github.com/nats-io/nats.go v1.11.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/nicksnyder/go-i18n/v2 v2.1.2
	github.com/nleeper/goment v1.4.1
	github.com/nsqio/go-nsq v1.0.8
	github.com/o1egl/paseto v1.0.0
	github.com/pelletier/go-toml v1.9.0 // indirect
	github.com/pquerna/otp v1.3.0
	github.com/prometheus/common v0.23.0 // indirect
	github.com/rs/zerolog v1.21.0
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/stripe/stripe-go/v72 v72.45.0
	github.com/tkuchiki/go-timezone v0.2.2 // indirect
	github.com/unrolled/secure v1.0.8
	github.com/wagslane/go-password-validator v0.3.0
	github.com/willf/bitset v1.1.11 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.19.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.19.0
	go.opentelemetry.io/otel v0.19.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.19.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.19.0
	go.opentelemetry.io/otel/metric v0.19.0
	go.opentelemetry.io/otel/sdk v0.19.0
	go.opentelemetry.io/otel/trace v0.19.0
	go.uber.org/ratelimit v0.2.0 // indirect
	gocloud.dev v0.23.0
	gocloud.dev/pubsub/kafkapubsub v0.23.0
	gocloud.dev/pubsub/natspubsub v0.23.0
	gocloud.dev/pubsub/rabbitpubsub v0.23.0
	gocloud.dev/secrets/hashivault v0.23.0
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c
	golang.org/x/sys v0.0.0-20210503173754-0981d6026fa6 // indirect
	golang.org/x/text v0.3.6
	google.golang.org/api v0.46.0 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/mikespook/gorbac.v2 v2.1.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
