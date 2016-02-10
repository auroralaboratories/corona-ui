// Package xinerama is the X client API for the XINERAMA extension.
package xinerama

// This file is automatically generated from xinerama.xml. Edit at your peril!

import (
	"github.com/BurntSushi/xgb"

	"github.com/BurntSushi/xgb/xproto"
)

// Init must be called before using the XINERAMA extension.
func Init(c *xgb.Conn) error {
	reply, err := xproto.QueryExtension(c, 8, "XINERAMA").Reply()
	switch {
	case err != nil:
		return err
	case !reply.Present:
		return xgb.Errorf("No extension named XINERAMA could be found on on the server.")
	}

	xgb.ExtLock.Lock()
	c.Extensions["XINERAMA"] = reply.MajorOpcode
	for evNum, fun := range xgb.NewExtEventFuncs["XINERAMA"] {
		xgb.NewEventFuncs[int(reply.FirstEvent)+evNum] = fun
	}
	for errNum, fun := range xgb.NewExtErrorFuncs["XINERAMA"] {
		xgb.NewErrorFuncs[int(reply.FirstError)+errNum] = fun
	}
	xgb.ExtLock.Unlock()

	return nil
}

func init() {
	xgb.NewExtEventFuncs["XINERAMA"] = make(map[int]xgb.NewEventFun)
	xgb.NewExtErrorFuncs["XINERAMA"] = make(map[int]xgb.NewErrorFun)
}

type ScreenInfo struct {
	XOrg   int16
	YOrg   int16
	Width  uint16
	Height uint16
}

// ScreenInfoRead reads a byte slice into a ScreenInfo value.
func ScreenInfoRead(buf []byte, v *ScreenInfo) int {
	b := 0

	v.XOrg = int16(xgb.Get16(buf[b:]))
	b += 2

	v.YOrg = int16(xgb.Get16(buf[b:]))
	b += 2

	v.Width = xgb.Get16(buf[b:])
	b += 2

	v.Height = xgb.Get16(buf[b:])
	b += 2

	return b
}

// ScreenInfoReadList reads a byte slice into a list of ScreenInfo values.
func ScreenInfoReadList(buf []byte, dest []ScreenInfo) int {
	b := 0
	for i := 0; i < len(dest); i++ {
		dest[i] = ScreenInfo{}
		b += ScreenInfoRead(buf[b:], &dest[i])
	}
	return xgb.Pad(b)
}

// Bytes writes a ScreenInfo value to a byte slice.
func (v ScreenInfo) Bytes() []byte {
	buf := make([]byte, 8)
	b := 0

	xgb.Put16(buf[b:], uint16(v.XOrg))
	b += 2

	xgb.Put16(buf[b:], uint16(v.YOrg))
	b += 2

	xgb.Put16(buf[b:], v.Width)
	b += 2

	xgb.Put16(buf[b:], v.Height)
	b += 2

	return buf[:b]
}

// ScreenInfoListBytes writes a list of ScreenInfo values to a byte slice.
func ScreenInfoListBytes(buf []byte, list []ScreenInfo) int {
	b := 0
	var structBytes []byte
	for _, item := range list {
		structBytes = item.Bytes()
		copy(buf[b:], structBytes)
		b += len(structBytes)
	}
	return xgb.Pad(b)
}

// Skipping definition for base type 'Bool'

// Skipping definition for base type 'Byte'

// Skipping definition for base type 'Card8'

// Skipping definition for base type 'Char'

// Skipping definition for base type 'Void'

// Skipping definition for base type 'Double'

// Skipping definition for base type 'Float'

// Skipping definition for base type 'Int16'

// Skipping definition for base type 'Int32'

// Skipping definition for base type 'Int8'

// Skipping definition for base type 'Card16'

// Skipping definition for base type 'Card32'

// GetScreenCountCookie is a cookie used only for GetScreenCount requests.
type GetScreenCountCookie struct {
	*xgb.Cookie
}

// GetScreenCount sends a checked request.
// If an error occurs, it will be returned with the reply by calling GetScreenCountCookie.Reply()
func GetScreenCount(c *xgb.Conn, Window xproto.Window) GetScreenCountCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'GetScreenCount' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(true, true)
	c.NewRequest(getScreenCountRequest(c, Window), cookie)
	return GetScreenCountCookie{cookie}
}

// GetScreenCountUnchecked sends an unchecked request.
// If an error occurs, it can only be retrieved using xgb.WaitForEvent or xgb.PollForEvent.
func GetScreenCountUnchecked(c *xgb.Conn, Window xproto.Window) GetScreenCountCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'GetScreenCount' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(false, true)
	c.NewRequest(getScreenCountRequest(c, Window), cookie)
	return GetScreenCountCookie{cookie}
}

// GetScreenCountReply represents the data returned from a GetScreenCount request.
type GetScreenCountReply struct {
	Sequence    uint16 // sequence number of the request for this reply
	Length      uint32 // number of bytes in this reply
	ScreenCount byte
	Window      xproto.Window
}

// Reply blocks and returns the reply data for a GetScreenCount request.
func (cook GetScreenCountCookie) Reply() (*GetScreenCountReply, error) {
	buf, err := cook.Cookie.Reply()
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, nil
	}
	return getScreenCountReply(buf), nil
}

// getScreenCountReply reads a byte slice into a GetScreenCountReply value.
func getScreenCountReply(buf []byte) *GetScreenCountReply {
	v := new(GetScreenCountReply)
	b := 1 // skip reply determinant

	v.ScreenCount = buf[b]
	b += 1

	v.Sequence = xgb.Get16(buf[b:])
	b += 2

	v.Length = xgb.Get32(buf[b:]) // 4-byte units
	b += 4

	v.Window = xproto.Window(xgb.Get32(buf[b:]))
	b += 4

	return v
}

// Write request to wire for GetScreenCount
// getScreenCountRequest writes a GetScreenCount request to a byte slice.
func getScreenCountRequest(c *xgb.Conn, Window xproto.Window) []byte {
	size := 8
	b := 0
	buf := make([]byte, size)

	buf[b] = c.Extensions["XINERAMA"]
	b += 1

	buf[b] = 2 // request opcode
	b += 1

	xgb.Put16(buf[b:], uint16(size/4)) // write request size in 4-byte units
	b += 2

	xgb.Put32(buf[b:], uint32(Window))
	b += 4

	return buf
}

// GetScreenSizeCookie is a cookie used only for GetScreenSize requests.
type GetScreenSizeCookie struct {
	*xgb.Cookie
}

// GetScreenSize sends a checked request.
// If an error occurs, it will be returned with the reply by calling GetScreenSizeCookie.Reply()
func GetScreenSize(c *xgb.Conn, Window xproto.Window, Screen uint32) GetScreenSizeCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'GetScreenSize' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(true, true)
	c.NewRequest(getScreenSizeRequest(c, Window, Screen), cookie)
	return GetScreenSizeCookie{cookie}
}

// GetScreenSizeUnchecked sends an unchecked request.
// If an error occurs, it can only be retrieved using xgb.WaitForEvent or xgb.PollForEvent.
func GetScreenSizeUnchecked(c *xgb.Conn, Window xproto.Window, Screen uint32) GetScreenSizeCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'GetScreenSize' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(false, true)
	c.NewRequest(getScreenSizeRequest(c, Window, Screen), cookie)
	return GetScreenSizeCookie{cookie}
}

// GetScreenSizeReply represents the data returned from a GetScreenSize request.
type GetScreenSizeReply struct {
	Sequence uint16 // sequence number of the request for this reply
	Length   uint32 // number of bytes in this reply
	// padding: 1 bytes
	Width  uint32
	Height uint32
	Window xproto.Window
	Screen uint32
}

// Reply blocks and returns the reply data for a GetScreenSize request.
func (cook GetScreenSizeCookie) Reply() (*GetScreenSizeReply, error) {
	buf, err := cook.Cookie.Reply()
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, nil
	}
	return getScreenSizeReply(buf), nil
}

// getScreenSizeReply reads a byte slice into a GetScreenSizeReply value.
func getScreenSizeReply(buf []byte) *GetScreenSizeReply {
	v := new(GetScreenSizeReply)
	b := 1 // skip reply determinant

	b += 1 // padding

	v.Sequence = xgb.Get16(buf[b:])
	b += 2

	v.Length = xgb.Get32(buf[b:]) // 4-byte units
	b += 4

	v.Width = xgb.Get32(buf[b:])
	b += 4

	v.Height = xgb.Get32(buf[b:])
	b += 4

	v.Window = xproto.Window(xgb.Get32(buf[b:]))
	b += 4

	v.Screen = xgb.Get32(buf[b:])
	b += 4

	return v
}

// Write request to wire for GetScreenSize
// getScreenSizeRequest writes a GetScreenSize request to a byte slice.
func getScreenSizeRequest(c *xgb.Conn, Window xproto.Window, Screen uint32) []byte {
	size := 12
	b := 0
	buf := make([]byte, size)

	buf[b] = c.Extensions["XINERAMA"]
	b += 1

	buf[b] = 3 // request opcode
	b += 1

	xgb.Put16(buf[b:], uint16(size/4)) // write request size in 4-byte units
	b += 2

	xgb.Put32(buf[b:], uint32(Window))
	b += 4

	xgb.Put32(buf[b:], Screen)
	b += 4

	return buf
}

// GetStateCookie is a cookie used only for GetState requests.
type GetStateCookie struct {
	*xgb.Cookie
}

// GetState sends a checked request.
// If an error occurs, it will be returned with the reply by calling GetStateCookie.Reply()
func GetState(c *xgb.Conn, Window xproto.Window) GetStateCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'GetState' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(true, true)
	c.NewRequest(getStateRequest(c, Window), cookie)
	return GetStateCookie{cookie}
}

// GetStateUnchecked sends an unchecked request.
// If an error occurs, it can only be retrieved using xgb.WaitForEvent or xgb.PollForEvent.
func GetStateUnchecked(c *xgb.Conn, Window xproto.Window) GetStateCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'GetState' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(false, true)
	c.NewRequest(getStateRequest(c, Window), cookie)
	return GetStateCookie{cookie}
}

// GetStateReply represents the data returned from a GetState request.
type GetStateReply struct {
	Sequence uint16 // sequence number of the request for this reply
	Length   uint32 // number of bytes in this reply
	State    byte
	Window   xproto.Window
}

// Reply blocks and returns the reply data for a GetState request.
func (cook GetStateCookie) Reply() (*GetStateReply, error) {
	buf, err := cook.Cookie.Reply()
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, nil
	}
	return getStateReply(buf), nil
}

// getStateReply reads a byte slice into a GetStateReply value.
func getStateReply(buf []byte) *GetStateReply {
	v := new(GetStateReply)
	b := 1 // skip reply determinant

	v.State = buf[b]
	b += 1

	v.Sequence = xgb.Get16(buf[b:])
	b += 2

	v.Length = xgb.Get32(buf[b:]) // 4-byte units
	b += 4

	v.Window = xproto.Window(xgb.Get32(buf[b:]))
	b += 4

	return v
}

// Write request to wire for GetState
// getStateRequest writes a GetState request to a byte slice.
func getStateRequest(c *xgb.Conn, Window xproto.Window) []byte {
	size := 8
	b := 0
	buf := make([]byte, size)

	buf[b] = c.Extensions["XINERAMA"]
	b += 1

	buf[b] = 1 // request opcode
	b += 1

	xgb.Put16(buf[b:], uint16(size/4)) // write request size in 4-byte units
	b += 2

	xgb.Put32(buf[b:], uint32(Window))
	b += 4

	return buf
}

// IsActiveCookie is a cookie used only for IsActive requests.
type IsActiveCookie struct {
	*xgb.Cookie
}

// IsActive sends a checked request.
// If an error occurs, it will be returned with the reply by calling IsActiveCookie.Reply()
func IsActive(c *xgb.Conn) IsActiveCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'IsActive' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(true, true)
	c.NewRequest(isActiveRequest(c), cookie)
	return IsActiveCookie{cookie}
}

// IsActiveUnchecked sends an unchecked request.
// If an error occurs, it can only be retrieved using xgb.WaitForEvent or xgb.PollForEvent.
func IsActiveUnchecked(c *xgb.Conn) IsActiveCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'IsActive' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(false, true)
	c.NewRequest(isActiveRequest(c), cookie)
	return IsActiveCookie{cookie}
}

// IsActiveReply represents the data returned from a IsActive request.
type IsActiveReply struct {
	Sequence uint16 // sequence number of the request for this reply
	Length   uint32 // number of bytes in this reply
	// padding: 1 bytes
	State uint32
}

// Reply blocks and returns the reply data for a IsActive request.
func (cook IsActiveCookie) Reply() (*IsActiveReply, error) {
	buf, err := cook.Cookie.Reply()
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, nil
	}
	return isActiveReply(buf), nil
}

// isActiveReply reads a byte slice into a IsActiveReply value.
func isActiveReply(buf []byte) *IsActiveReply {
	v := new(IsActiveReply)
	b := 1 // skip reply determinant

	b += 1 // padding

	v.Sequence = xgb.Get16(buf[b:])
	b += 2

	v.Length = xgb.Get32(buf[b:]) // 4-byte units
	b += 4

	v.State = xgb.Get32(buf[b:])
	b += 4

	return v
}

// Write request to wire for IsActive
// isActiveRequest writes a IsActive request to a byte slice.
func isActiveRequest(c *xgb.Conn) []byte {
	size := 4
	b := 0
	buf := make([]byte, size)

	buf[b] = c.Extensions["XINERAMA"]
	b += 1

	buf[b] = 4 // request opcode
	b += 1

	xgb.Put16(buf[b:], uint16(size/4)) // write request size in 4-byte units
	b += 2

	return buf
}

// QueryScreensCookie is a cookie used only for QueryScreens requests.
type QueryScreensCookie struct {
	*xgb.Cookie
}

// QueryScreens sends a checked request.
// If an error occurs, it will be returned with the reply by calling QueryScreensCookie.Reply()
func QueryScreens(c *xgb.Conn) QueryScreensCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'QueryScreens' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(true, true)
	c.NewRequest(queryScreensRequest(c), cookie)
	return QueryScreensCookie{cookie}
}

// QueryScreensUnchecked sends an unchecked request.
// If an error occurs, it can only be retrieved using xgb.WaitForEvent or xgb.PollForEvent.
func QueryScreensUnchecked(c *xgb.Conn) QueryScreensCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'QueryScreens' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(false, true)
	c.NewRequest(queryScreensRequest(c), cookie)
	return QueryScreensCookie{cookie}
}

// QueryScreensReply represents the data returned from a QueryScreens request.
type QueryScreensReply struct {
	Sequence uint16 // sequence number of the request for this reply
	Length   uint32 // number of bytes in this reply
	// padding: 1 bytes
	Number uint32
	// padding: 20 bytes
	ScreenInfo []ScreenInfo // size: xgb.Pad((int(Number) * 8))
}

// Reply blocks and returns the reply data for a QueryScreens request.
func (cook QueryScreensCookie) Reply() (*QueryScreensReply, error) {
	buf, err := cook.Cookie.Reply()
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, nil
	}
	return queryScreensReply(buf), nil
}

// queryScreensReply reads a byte slice into a QueryScreensReply value.
func queryScreensReply(buf []byte) *QueryScreensReply {
	v := new(QueryScreensReply)
	b := 1 // skip reply determinant

	b += 1 // padding

	v.Sequence = xgb.Get16(buf[b:])
	b += 2

	v.Length = xgb.Get32(buf[b:]) // 4-byte units
	b += 4

	v.Number = xgb.Get32(buf[b:])
	b += 4

	b += 20 // padding

	v.ScreenInfo = make([]ScreenInfo, v.Number)
	b += ScreenInfoReadList(buf[b:], v.ScreenInfo)

	return v
}

// Write request to wire for QueryScreens
// queryScreensRequest writes a QueryScreens request to a byte slice.
func queryScreensRequest(c *xgb.Conn) []byte {
	size := 4
	b := 0
	buf := make([]byte, size)

	buf[b] = c.Extensions["XINERAMA"]
	b += 1

	buf[b] = 5 // request opcode
	b += 1

	xgb.Put16(buf[b:], uint16(size/4)) // write request size in 4-byte units
	b += 2

	return buf
}

// QueryVersionCookie is a cookie used only for QueryVersion requests.
type QueryVersionCookie struct {
	*xgb.Cookie
}

// QueryVersion sends a checked request.
// If an error occurs, it will be returned with the reply by calling QueryVersionCookie.Reply()
func QueryVersion(c *xgb.Conn, Major byte, Minor byte) QueryVersionCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'QueryVersion' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(true, true)
	c.NewRequest(queryVersionRequest(c, Major, Minor), cookie)
	return QueryVersionCookie{cookie}
}

// QueryVersionUnchecked sends an unchecked request.
// If an error occurs, it can only be retrieved using xgb.WaitForEvent or xgb.PollForEvent.
func QueryVersionUnchecked(c *xgb.Conn, Major byte, Minor byte) QueryVersionCookie {
	if _, ok := c.Extensions["XINERAMA"]; !ok {
		panic("Cannot issue request 'QueryVersion' using the uninitialized extension 'XINERAMA'. xinerama.Init(connObj) must be called first.")
	}
	cookie := c.NewCookie(false, true)
	c.NewRequest(queryVersionRequest(c, Major, Minor), cookie)
	return QueryVersionCookie{cookie}
}

// QueryVersionReply represents the data returned from a QueryVersion request.
type QueryVersionReply struct {
	Sequence uint16 // sequence number of the request for this reply
	Length   uint32 // number of bytes in this reply
	// padding: 1 bytes
	Major uint16
	Minor uint16
}

// Reply blocks and returns the reply data for a QueryVersion request.
func (cook QueryVersionCookie) Reply() (*QueryVersionReply, error) {
	buf, err := cook.Cookie.Reply()
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, nil
	}
	return queryVersionReply(buf), nil
}

// queryVersionReply reads a byte slice into a QueryVersionReply value.
func queryVersionReply(buf []byte) *QueryVersionReply {
	v := new(QueryVersionReply)
	b := 1 // skip reply determinant

	b += 1 // padding

	v.Sequence = xgb.Get16(buf[b:])
	b += 2

	v.Length = xgb.Get32(buf[b:]) // 4-byte units
	b += 4

	v.Major = xgb.Get16(buf[b:])
	b += 2

	v.Minor = xgb.Get16(buf[b:])
	b += 2

	return v
}

// Write request to wire for QueryVersion
// queryVersionRequest writes a QueryVersion request to a byte slice.
func queryVersionRequest(c *xgb.Conn, Major byte, Minor byte) []byte {
	size := 8
	b := 0
	buf := make([]byte, size)

	buf[b] = c.Extensions["XINERAMA"]
	b += 1

	buf[b] = 0 // request opcode
	b += 1

	xgb.Put16(buf[b:], uint16(size/4)) // write request size in 4-byte units
	b += 2

	buf[b] = Major
	b += 1

	buf[b] = Minor
	b += 1

	return buf
}
