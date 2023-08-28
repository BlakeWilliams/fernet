package monte

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/fernet"
	"github.com/stretchr/testify/require"
)

type Team struct {
	Name string
}

type TeamController struct {
	team *Team
	// *Controller[int, string]
}

func (c *TeamController) Init(res fernet.Response, req *fernet.Request[int], next fernet.Handler[int]) {
	c.team = &Team{Name: req.Param("team")}
	next(res, req)
}

func (c *TeamController) Routes(parentRouter fernet.Routable[int]) {
	parentRouter.Get("/show/:team", c.Show)
}

func (c *TeamController) Show(res fernet.Response, req *fernet.Request[int]) {
	res.Write([]byte("hello " + c.team.Name))
}

func TestController(t *testing.T) {
	router := fernet.New[int]()

	router.Register(NewController[int](&TeamController{}))

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/show/my_team", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, 200, res.Code)
	require.Equal(t, "hello my_team", res.Body.String())
}
