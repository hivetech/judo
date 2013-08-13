package loggo_test

import (
	"fmt"
	"time"

	. "launchpad.net/gocheck"
	"launchpad.net/loggo"
)

type writerBasicsSuite struct{}

var _ = Suite(&writerBasicsSuite{})

func (s *writerBasicsSuite) TearDownTest(c *C) {
	loggo.ResetWriters()
}

func (*writerBasicsSuite) TestRemoveDefaultWriter(c *C) {
	defaultWriter, level, err := loggo.RemoveWriter("default")
	c.Assert(err, IsNil)
	c.Assert(level, Equals, loggo.TRACE)
	c.Assert(defaultWriter, Not(IsNil))

	// Trying again fails.
	defaultWriter, level, err = loggo.RemoveWriter("default")
	c.Assert(err, ErrorMatches, `Writer "default" is not registered`)
	c.Assert(level, Equals, loggo.UNSPECIFIED)
	c.Assert(defaultWriter, IsNil)
}

func (*writerBasicsSuite) TestRegisterWriterExistingName(c *C) {
	err := loggo.RegisterWriter("default", &loggo.TestWriter{}, loggo.INFO)
	c.Assert(err, ErrorMatches, `there is already a Writer registered with the name "default"`)
}

func (*writerBasicsSuite) TestRegisterNilWriter(c *C) {
	err := loggo.RegisterWriter("nil", nil, loggo.INFO)
	c.Assert(err, ErrorMatches, `Writer cannot be nil`)
}

func (*writerBasicsSuite) TestRegisterWriterTypedNil(c *C) {
	// If the interface is a typed nil, we have to trust the user.
	var writer *loggo.TestWriter
	err := loggo.RegisterWriter("nil", writer, loggo.INFO)
	c.Assert(err, IsNil)
}

func (*writerBasicsSuite) TestReplaceDefaultWriter(c *C) {
	oldWriter, err := loggo.ReplaceDefaultWriter(&loggo.TestWriter{})
	c.Assert(oldWriter, NotNil)
	c.Assert(err, IsNil)
}

func (*writerBasicsSuite) TestReplaceDefaultWriterWithNil(c *C) {
	oldWriter, err := loggo.ReplaceDefaultWriter(nil)
	c.Assert(oldWriter, IsNil)
	c.Assert(err, ErrorMatches, "Writer cannot be nil")
}

func (*writerBasicsSuite) TestReplaceDefaultWriterNoDefault(c *C) {
	loggo.RemoveWriter("default")
	oldWriter, err := loggo.ReplaceDefaultWriter(&loggo.TestWriter{})
	c.Assert(oldWriter, IsNil)
	c.Assert(err, ErrorMatches, `there is no "default" writer`)
}

func (s *writerBasicsSuite) TestWillWrite(c *C) {
	// By default, the root logger watches TRACE messages
	c.Assert(loggo.WillWrite(loggo.TRACE), Equals, true)
	// Note: ReplaceDefaultWriter doesn't let us change the default log
	//	 level :(
	writer, _, err := loggo.RemoveWriter("default")
	c.Assert(err, IsNil)
	c.Assert(writer, NotNil)
	err = loggo.RegisterWriter("default", writer, loggo.CRITICAL)
	c.Assert(err, IsNil)
	c.Assert(loggo.WillWrite(loggo.TRACE), Equals, false)
	c.Assert(loggo.WillWrite(loggo.DEBUG), Equals, false)
	c.Assert(loggo.WillWrite(loggo.INFO), Equals, false)
	c.Assert(loggo.WillWrite(loggo.WARNING), Equals, false)
	c.Assert(loggo.WillWrite(loggo.CRITICAL), Equals, true)
}

type writerSuite struct {
	logger loggo.Logger
}

var _ = Suite(&writerSuite{})

func (s *writerSuite) SetUpTest(c *C) {
	loggo.ResetLoggers()
	loggo.RemoveWriter("default")
	s.logger = loggo.GetLogger("test.writer")
	// Make it so the logger itself writes all messages.
	s.logger.SetLogLevel(loggo.TRACE)
}

func (s *writerSuite) TearDownTest(c *C) {
	loggo.ResetWriters()
}

func (s *writerSuite) TearDownSuite(c *C) {
	loggo.ResetLoggers()
}

func (s *writerSuite) TestWritingCapturesFileAndLineAndModule(c *C) {
	writer := &loggo.TestWriter{}
	err := loggo.RegisterWriter("test", writer, loggo.INFO)
	c.Assert(err, IsNil)

	s.logger.Infof("Info message")

	// WARNING: test checks the line number of the above logger lines, this
	// will mean that if the above line moves, the test will fail unless
	// updated.
	c.Assert(writer.Log, HasLen, 1)
	c.Assert(writer.Log[0].Filename, Equals, "writer_test.go")
	c.Assert(writer.Log[0].Line, Equals, 112)
	c.Assert(writer.Log[0].Module, Equals, "test.writer")
}

func (s *writerSuite) TestWritingLimitWarning(c *C) {
	writer := &loggo.TestWriter{}
	err := loggo.RegisterWriter("test", writer, loggo.WARNING)
	c.Assert(err, IsNil)

	start := time.Now()
	s.logger.Criticalf("Something critical.")
	s.logger.Errorf("An error.")
	s.logger.Warningf("A warning message")
	s.logger.Infof("Info message")
	s.logger.Tracef("Trace the function")
	end := time.Now()

	c.Assert(writer.Log, HasLen, 3)
	c.Assert(writer.Log[0].Level, Equals, loggo.CRITICAL)
	c.Assert(writer.Log[0].Message, Equals, "Something critical.")
	c.Assert(writer.Log[0].Timestamp, Between(start, end))

	c.Assert(writer.Log[1].Level, Equals, loggo.ERROR)
	c.Assert(writer.Log[1].Message, Equals, "An error.")
	c.Assert(writer.Log[1].Timestamp, Between(start, end))

	c.Assert(writer.Log[2].Level, Equals, loggo.WARNING)
	c.Assert(writer.Log[2].Message, Equals, "A warning message")
	c.Assert(writer.Log[2].Timestamp, Between(start, end))
}

func (s *writerSuite) TestWritingLimitTrace(c *C) {
	writer := &loggo.TestWriter{}
	err := loggo.RegisterWriter("test", writer, loggo.TRACE)
	c.Assert(err, IsNil)

	start := time.Now()
	s.logger.Criticalf("Something critical.")
	s.logger.Errorf("An error.")
	s.logger.Warningf("A warning message")
	s.logger.Infof("Info message")
	s.logger.Tracef("Trace the function")
	end := time.Now()

	c.Assert(writer.Log, HasLen, 5)
	c.Assert(writer.Log[0].Level, Equals, loggo.CRITICAL)
	c.Assert(writer.Log[0].Message, Equals, "Something critical.")
	c.Assert(writer.Log[0].Timestamp, Between(start, end))

	c.Assert(writer.Log[1].Level, Equals, loggo.ERROR)
	c.Assert(writer.Log[1].Message, Equals, "An error.")
	c.Assert(writer.Log[1].Timestamp, Between(start, end))

	c.Assert(writer.Log[2].Level, Equals, loggo.WARNING)
	c.Assert(writer.Log[2].Message, Equals, "A warning message")
	c.Assert(writer.Log[2].Timestamp, Between(start, end))

	c.Assert(writer.Log[3].Level, Equals, loggo.INFO)
	c.Assert(writer.Log[3].Message, Equals, "Info message")
	c.Assert(writer.Log[3].Timestamp, Between(start, end))

	c.Assert(writer.Log[4].Level, Equals, loggo.TRACE)
	c.Assert(writer.Log[4].Message, Equals, "Trace the function")
	c.Assert(writer.Log[4].Timestamp, Between(start, end))
}

func (s *writerSuite) TestMultipleWriters(c *C) {
	errorWriter := &loggo.TestWriter{}
	err := loggo.RegisterWriter("error", errorWriter, loggo.ERROR)
	c.Assert(err, IsNil)
	warningWriter := &loggo.TestWriter{}
	err = loggo.RegisterWriter("warning", warningWriter, loggo.WARNING)
	c.Assert(err, IsNil)
	infoWriter := &loggo.TestWriter{}
	err = loggo.RegisterWriter("info", infoWriter, loggo.INFO)
	c.Assert(err, IsNil)
	traceWriter := &loggo.TestWriter{}
	err = loggo.RegisterWriter("trace", traceWriter, loggo.TRACE)
	c.Assert(err, IsNil)

	s.logger.Errorf("An error.")
	s.logger.Warningf("A warning message")
	s.logger.Infof("Info message")
	s.logger.Tracef("Trace the function")

	c.Assert(errorWriter.Log, HasLen, 1)
	c.Assert(warningWriter.Log, HasLen, 2)
	c.Assert(infoWriter.Log, HasLen, 3)
	c.Assert(traceWriter.Log, HasLen, 4)
}

func Between(start, end time.Time) Checker {
	if end.Before(start) {
		return &betweenChecker{end, start}
	}
	return &betweenChecker{start, end}
}

type betweenChecker struct {
	start, end time.Time
}

func (checker *betweenChecker) Info() *CheckerInfo {
	info := CheckerInfo{
		Name:   "Between",
		Params: []string{"obtained"},
	}
	return &info
}

func (checker *betweenChecker) Check(params []interface{}, names []string) (result bool, error string) {
	when, ok := params[0].(time.Time)
	if !ok {
		return false, "obtained value type must be time.Time"
	}
	if when.Before(checker.start) {
		return false, fmt.Sprintf("obtained value %#v type must before start value of %#v", when, checker.start)
	}
	if when.After(checker.end) {
		return false, fmt.Sprintf("obtained value %#v type must after end value of %#v", when, checker.end)
	}
	return true, ""
}
