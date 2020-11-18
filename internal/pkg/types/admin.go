package types

import (
	"context"
	"net/http"
)

type (
	// AdminServer describes a structure capable of serving traffic related to users.
	AdminServer interface {
		BanHandler(res http.ResponseWriter, req *http.Request)
	}

	// AdminAuditManager describes a structure capable of managing audit entries for admin events.
	AdminAuditManager interface {
		LogUserBanEvent(ctx context.Context, banGiver, banReceiver uint64)
	}
)
