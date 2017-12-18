package tests

import (
	"github.com/revel/revel/testing"
	"schneidernet/smarthome/app/dao"
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
	x := []byte("payload")
	dao.SaveToken("code", &x)
	y := []byte("payload1")
	dao.SaveToken("code1", &y)

	t.AssertEqual(*dao.GetToken("code"), "payload")
	t.AssertNotEqual(*dao.GetToken("code"), "payload1")
	t.AssertEqual(*dao.GetToken("code1"), "payload1")
	t.Assert(dao.GetToken("code2") == nil)

	t.Assert(dao.DeleteToken("code1") == nil)
	t.Assert(dao.DeleteToken("code1") != nil)

}
