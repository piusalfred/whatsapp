module github.com/piusalfred/whatsapp

go 1.24

require (
	github.com/google/go-cmp v0.7.0
	go.uber.org/mock v0.5.0
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Ladicle/tabwriter v1.0.0 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v1.1.3 // indirect
	github.com/alecthomas/assert/v2 v2.11.0 // indirect
	github.com/alecthomas/chroma/v2 v2.14.0 // indirect
	github.com/chainguard-dev/git-urls v1.0.2 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/cyphar/filepath-securejoin v0.3.6 // indirect
	github.com/daixiang0/gci v0.13.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/dominikbraun/graph v0.23.0 // indirect
	github.com/elliotchance/orderedmap/v2 v2.7.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.1 // indirect
	github.com/go-git/go-git/v5 v5.13.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-task/task/v3 v3.41.0 // indirect
	github.com/go-task/template v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-zglob v0.0.6 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/onsi/gomega v1.36.2 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sajari/fuzzy v1.0.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.3.0 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.7.0 // indirect
	mvdan.cc/sh/v3 v3.10.0 // indirect
)

tool (
	github.com/daixiang0/gci
	github.com/go-task/task/v3/cmd/task
	github.com/piusalfred/whatsapp
	github.com/piusalfred/whatsapp/auth
	github.com/piusalfred/whatsapp/business
	github.com/piusalfred/whatsapp/business/analytics
	github.com/piusalfred/whatsapp/config
	github.com/piusalfred/whatsapp/conversation/automation
	github.com/piusalfred/whatsapp/examples/block
	github.com/piusalfred/whatsapp/examples/webhooks/listeners
	github.com/piusalfred/whatsapp/examples/webhooks/simple
	github.com/piusalfred/whatsapp/flow
	github.com/piusalfred/whatsapp/media
	github.com/piusalfred/whatsapp/message
	github.com/piusalfred/whatsapp/mocks/auth
	github.com/piusalfred/whatsapp/mocks/business
	github.com/piusalfred/whatsapp/mocks/business/analytics
	github.com/piusalfred/whatsapp/mocks/config
	github.com/piusalfred/whatsapp/mocks/conversation/automation
	github.com/piusalfred/whatsapp/mocks/flow
	github.com/piusalfred/whatsapp/mocks/http
	github.com/piusalfred/whatsapp/mocks/media
	github.com/piusalfred/whatsapp/mocks/message
	github.com/piusalfred/whatsapp/mocks/phonenumber
	github.com/piusalfred/whatsapp/mocks/qrcode
	github.com/piusalfred/whatsapp/mocks/user
	github.com/piusalfred/whatsapp/phonenumber
	github.com/piusalfred/whatsapp/pkg/crypto
	github.com/piusalfred/whatsapp/pkg/errors
	github.com/piusalfred/whatsapp/pkg/http
	github.com/piusalfred/whatsapp/pkg/types
	github.com/piusalfred/whatsapp/qrcode
	github.com/piusalfred/whatsapp/user
	github.com/piusalfred/whatsapp/webhooks
	github.com/piusalfred/whatsapp/webhooks/business
	github.com/piusalfred/whatsapp/webhooks/flow
	github.com/piusalfred/whatsapp/webhooks/message
	go.uber.org/mock/mockgen
	mvdan.cc/gofumpt
)
