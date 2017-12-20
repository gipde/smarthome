package tests

import (
	"github.com/ory/fosite"
	"github.com/revel/revel/testing"
	"schneidernet/smarthome/app/dao"
	"time"
)

type DaoTest struct {
	testing.TestSuite
}

func (t *DaoTest) Before() {
	println("Set up")
}

func (t *DaoTest) After() {
	println("Tear down")
}

func (t *DaoTest) TestExample() {
	x := []byte("payload1")
	dao.SaveToken("signatur1", "tokenid1", fosite.AccessToken, time.Now().UTC().Add(time.Hour), &x)
	y := []byte("payload2")
	dao.SaveToken("signatur2", "tokenid2", fosite.RefreshToken, time.Now().UTC().Add(time.Hour), &y)

	t.AssertEqual(*dao.GetTokenBySignature("signatur1"), "payload1")
	t.AssertNotEqual(*dao.GetTokenBySignature("signatur2"), "payload1")
	t.AssertEqual(*dao.GetTokenBySignature("signatur1"), "payload1")
	t.Assert(dao.GetTokenBySignature("signatur3") == nil)
	t.AssertEqual(*dao.GetTokenByTokenID("tokenid1", fosite.AccessToken), "payload1")
	t.Assert(dao.GetTokenByTokenID("tokenid1", fosite.RefreshToken) == nil)

	t.Assert(dao.DeleteToken("signatur1") == nil)
	t.Assert(dao.DeleteToken("signatur1") != nil)

}
