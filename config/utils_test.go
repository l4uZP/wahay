package config

import (
	"crypto/rand"
	"errors"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
	. "gopkg.in/check.v1"
	"io"
	"net"
	"os"
	"strings"
)

type UtilsSuite struct{}

var _ = Suite(&UtilsSuite{})

func (u *UtilsSuite) Test_RandomString_returnsARandomStringIfTheReaderHaveEnoughData(c *C) {
	defer gostub.New().Stub(&rand.Reader, strings.NewReader("hello")).Reset()

	dest := new([5]byte)
	res := RandomString(dest[:])
	c.Assert(res, IsNil)
	c.Assert(string(dest[:]), Equals, "68656")
}

type errReader struct {
	e error
}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, r.e
}

func (u *UtilsSuite) Test_RandomString_returnsAnErrorIfTheReaderGivesAnError(c *C) {
	defer gostub.New().Stub(&rand.Reader, &errReader{io.EOF}).Reset()

	dest := new([5]byte)
	res := RandomString(dest[:])
	c.Assert(res, ErrorMatches, "EOF")
}
func (u *UtilsSuite) Test_RandomString_returnsAnErrorIfTheReaderDoesntHaveEnoughData(c *C) {
	defer gostub.New().Stub(&rand.Reader, strings.NewReader("short")).Reset()

	dest := new([10]byte)
	res := RandomString(dest[:])
	c.Assert(res, ErrorMatches, "unexpected EOF")
}

func (u *UtilsSuite) Test_WithHome_returnsTheHomeOfTheHostConcatenatedWithTheGivenPath(c *C) {
	defer gostub.New().SetEnv("HOME", "/my/custom/home").Reset()

	c.Assert(WithHome("hello/goodbye.txt"), Equals, "/my/custom/home/hello/goodbye.txt")

	_ = os.Setenv("HOME", "/another/custom/home")
	c.Assert(WithHome("something else/bla/root//foo.ext.jpg"), Equals, "/another/custom/home/something else/bla/root/foo.ext.jpg")
}

func (u *UtilsSuite) Test_XdgConfigHome_returnsTheCustomEnvironmentVariableDefinedInOrIfStandardEnvIsNotPresent(c *C) {
	defer gostub.New().SetEnv("XDG_CONFIG_HOME", "").Reset()
	defer gostub.New().SetEnv("HOME", "/a/custom/home").Reset()

	c.Assert(XdgConfigHome(), Equals, "/a/custom/home/.config")
}

func (u *UtilsSuite) Test_XdgConfigHome_returnsTheStandardEnvIfItIsPresent(c *C) {
	defer gostub.New().SetEnv("XDG_CONFIG_HOME", "/a/config/standard/directory").Reset()

	c.Assert(xdgOrWithHome("XDG_CONFIG_HOME", "/custom/config/directory"), Equals, "/a/config/standard/directory")
}

func (u *UtilsSuite) Test_XdgDataHome_returnsTheCustomEnvironmentVariableDefinedInOrIfStandardEnvIsNotPresent(c *C) {
	defer gostub.New().SetEnv("XDG_DATA_HOME", "").Reset()
	defer gostub.New().SetEnv("HOME", "/a/custom/home").Reset()

	c.Assert(XdgConfigHome(), Equals, "/a/custom/home/.config")
}

func (u *UtilsSuite) Test_XdgDataHome_returnsTheStandardEnvIfItIsPresent(c *C) {
	defer gostub.New().SetEnv("XDG_DATA_HOME", "/a/config/standard/directory").Reset()

	c.Assert(xdgOrWithHome("XDG_DATA_HOME", "/custom/config/directory"), Equals, "/a/config/standard/directory")
}

type mockListener struct {
	mock.Mock
}

func (l *mockListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (l *mockListener) Close() error {
	ret := l.Called()
	return ret.Error(0)
}

func (l *mockListener) Addr() net.Addr {
	return nil
}

type mockNet struct {
	mock.Mock
}

func (m *mockNet) Listen(network, dir string) (net.Listener, error) {
	ret := m.Called(network, dir)
	return ret.Get(0).(net.Listener), ret.Error(1)
}

func (u *UtilsSuite) Test_IsPortAvailable_returnsTrueIfThePortIsAvailable(c *C) {
	ml := &mockListener{}
	mn := &mockNet{}

	defer gostub.New().Stub(&listen, mn.Listen).Reset()

	mn.On("Listen", "tcp", ":10001").Return(ml, nil).Once()
	ml.On("Close").Return(nil)

	c.Assert(IsPortAvailable(10001), Equals, true)

	mn.AssertExpectations(c)
	ml.AssertExpectations(c)
}

func (u *UtilsSuite) Test_IsPortAvailable_returnsTrueIfAnotherPortIsAvailable(c *C) {
	ml := &mockListener{}

	defer gostub.New().Stub(&listen, func(net string, dir string) (net.Listener, error) {
		c.Assert(net, Equals, "tcp")
		c.Assert(dir, Equals, ":10002")
		return ml, nil
	}).Reset()

	ml.On("Close").Return(nil)

	c.Assert(IsPortAvailable(10002), Equals, true)
}

type mockRandom struct {
	mock.Mock
}

func (m *mockRandom) Int31n(v int32) int32 {
	return int32(m.Called(v).Int(0))
}

func (u *UtilsSuite) Test_GetRandomPort_returnsThePortAvailableBetweenSomePorts(c *C) {
	mr := &mockRandom{}
	mn := &mockNet{}

	defer gostub.New().Stub(&randomInt31, mr.Int31n).Reset()
	defer gostub.New().Stub(&listen, mn.Listen).Reset()

	mr.On("Int31n", int32(50000)).Return(2530).Once()
	mr.On("Int31n", int32(50000)).Return(5679).Once()

	ml := &mockListener{}

	mn.On("Listen", "tcp", ":12530").Return(ml, errors.New("error")).Once()
	mn.On("Listen", "tcp", ":15679").Return(ml, nil).Once()

	ml.On("Close").Return(nil).Once()

	c.Assert(GetRandomPort(), Equals, 15679)
}

func (u *UtilsSuite) Test_IsPortAvailable_returnsFalseIfThePortIsNotAvailable(c *C) {
	defer gostub.New().StubFunc(&listen, nil, errors.New("port already taken")).Reset()

	c.Assert(IsPortAvailable(5555), Equals, false)
}

func (u *UtilsSuite) Test_IsPortAvailable_returnsFalseIfThePortWasAvailableButSomethingWentWrongWhenTestingIt(c *C) {
	ml := &mockListener{}
	defer gostub.New().StubFunc(&listen, ml, nil).Reset()
	ml.On("Close").Return(errors.New("oh no")).Once()

	c.Assert(IsPortAvailable(65501), Equals, false)
}

func (u *UtilsSuite) Test_RandomPort_ReturnsAPortBetween10000And59999(c *C) {
	defer gostub.New().StubFunc(&randomInt31, int32(0)).Reset()
	c.Assert(RandomPort(), Equals, 10000)

	defer gostub.New().StubFunc(&randomInt31, int32(25000)).Reset()
	c.Assert(RandomPort(), Equals, 35000)

	defer gostub.New().StubFunc(&randomInt31, int32(49999)).Reset()
	c.Assert(RandomPort(), Equals, 59999)
}

func (u *UtilsSuite) Test_CheckPort_ReturnsFalseIfTheGivenValueIsNegative(c *C) {
	c.Assert(CheckPort(-1), Equals, false)
	c.Assert(CheckPort(0), Equals, false)
	c.Assert(CheckPort(65536), Equals, false)
}

func (u *UtilsSuite) Test_CheckPort_ReturnsTrueIfTheGivenValueIsBetweenOneAnd65535(c *C) {
	c.Assert(CheckPort(1), Equals, true)
	c.Assert(CheckPort(90), Equals, true)
	c.Assert(CheckPort(65535), Equals, true)
}

func (u *UtilsSuite) Test_DetectLanguage_returnsEngIfCannotDetectTheSystemLanguage(c *C) {
	defer gostub.New().StubFunc(&detectLanguage, language.Und, nil).Reset()

	c.Assert(DetectLanguage(), Equals, language.English)
}

func (u *UtilsSuite) Test_DetectLanguage_returnsTheLanguageDetected(c *C) {
	defer gostub.New().StubFunc(&detectLanguage, language.Hindi, nil).Reset()

	c.Assert(DetectLanguage(), Equals, language.Hindi)
}
func (u *UtilsSuite) Test_DetectLanguage_returnsEnglishIfEnglishIsDetected(c *C) {
	defer gostub.New().StubFunc(&detectLanguage, language.English, nil).Reset()

	c.Assert(DetectLanguage(), Equals, language.English)
}