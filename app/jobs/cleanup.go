
package jobs

import (
	"github.com/robfig/revel"
	"github.com/robfig/revel/modules/jobs/app/jobs"
	"github.com/PacketFire/goqdb/app/controllers"
	"github.com/PacketFire/goqdb/app/models"
	"time"
)

// midnight
var threshold = time.Now().Truncate(24 * time.Hour).Unix()

func init () {
	revel.OnAppStart(func () {
		jobs.Schedule("@midnight", jobs.Func(cleanup))
	})
}

func cleanup () {
	_, err := controllers.Dbm.Exec(`DELETE FROM VoteEntry WHERE Created < ? AND VoteType != ?`,
		threshold, models.VOTE_DELETE)

	if err != nil {
		panic(err)
	}
}
