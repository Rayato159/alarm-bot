package atomikkuTest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Rayato159/awaken-discord-bot/config"
	"github.com/Rayato159/awaken-discord-bot/pkg/kawaiiauth"
	"github.com/stretchr/testify/assert"
)

type testSignToken struct {
	label string
	req   *kawaiiauth.MiniPayload
}

func TestSignToken(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewConfig("../.env")

	tests := []testSignToken{
		{
			label: "success",
			req:   &kawaiiauth.MiniPayload{},
		},
	}

	for _, test := range tests {
		fmt.Println(test.label)

		auth := kawaiiauth.NewKawaiiAuth(cfg)
		token, err := auth.SignJwtToken(ctx, 8640000, test.req)
		assert.Empty(t, err)
		assert.NotEmpty(t, token)

		err = os.WriteFile(fmt.Sprintf("./access_token_%s.txt", test.label), []byte(token), 0777)
		assert.Empty(t, err)
	}
}
