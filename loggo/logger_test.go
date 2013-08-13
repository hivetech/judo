package loggo_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "launchpad.net/gocheck"
	"launchpad.net/loggo"
)

func Test(t *testing.T) {
	TestingT(t)
}

type loggerSuite struct{}

var _ = Suite(&loggerSuite{})

func (*loggerSuite) SetUpTest(c *C) {
	loggo.ResetLoggers()
}

func (*loggerSuite) TestRootLogger(c *C) {
	root := loggo.Logger{}
	c.Assert(root.Name(), Equals, "<root>")
	c.Assert(root.IsErrorEnabled(), Equals, true)
	c.Assert(root.IsWarningEnabled(), Equals, true)
	c.Assert(root.IsInfoEnabled(), Equals, false)
	c.Assert(root.IsDebugEnabled(), Equals, false)
	c.Assert(root.IsTraceEnabled(), Equals, false)
}

func (*loggerSuite) TestModuleName(c *C) {
	logger := loggo.GetLogger("loggo.testing")
	c.Assert(logger.Name(), Equals, "loggo.testing")
}

func (*loggerSuite) TestSetLevel(c *C) {
	logger := loggo.GetLogger("testing")

	c.Assert(logger.LogLevel(), Equals, loggo.UNSPECIFIED)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.WARNING)
	c.Assert(logger.IsErrorEnabled(), Equals, true)
	c.Assert(logger.IsWarningEnabled(), Equals, true)
	c.Assert(logger.IsInfoEnabled(), Equals, false)
	c.Assert(logger.IsDebugEnabled(), Equals, false)
	c.Assert(logger.IsTraceEnabled(), Equals, false)
	logger.SetLogLevel(loggo.TRACE)
	c.Assert(logger.LogLevel(), Equals, loggo.TRACE)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.TRACE)
	c.Assert(logger.IsErrorEnabled(), Equals, true)
	c.Assert(logger.IsWarningEnabled(), Equals, true)
	c.Assert(logger.IsInfoEnabled(), Equals, true)
	c.Assert(logger.IsDebugEnabled(), Equals, true)
	c.Assert(logger.IsTraceEnabled(), Equals, true)
	logger.SetLogLevel(loggo.DEBUG)
	c.Assert(logger.LogLevel(), Equals, loggo.DEBUG)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.DEBUG)
	c.Assert(logger.IsErrorEnabled(), Equals, true)
	c.Assert(logger.IsWarningEnabled(), Equals, true)
	c.Assert(logger.IsInfoEnabled(), Equals, true)
	c.Assert(logger.IsDebugEnabled(), Equals, true)
	c.Assert(logger.IsTraceEnabled(), Equals, false)
	logger.SetLogLevel(loggo.INFO)
	c.Assert(logger.LogLevel(), Equals, loggo.INFO)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.INFO)
	c.Assert(logger.IsErrorEnabled(), Equals, true)
	c.Assert(logger.IsWarningEnabled(), Equals, true)
	c.Assert(logger.IsInfoEnabled(), Equals, true)
	c.Assert(logger.IsDebugEnabled(), Equals, false)
	c.Assert(logger.IsTraceEnabled(), Equals, false)
	logger.SetLogLevel(loggo.WARNING)
	c.Assert(logger.LogLevel(), Equals, loggo.WARNING)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.WARNING)
	c.Assert(logger.IsErrorEnabled(), Equals, true)
	c.Assert(logger.IsWarningEnabled(), Equals, true)
	c.Assert(logger.IsInfoEnabled(), Equals, false)
	c.Assert(logger.IsDebugEnabled(), Equals, false)
	c.Assert(logger.IsTraceEnabled(), Equals, false)
	logger.SetLogLevel(loggo.ERROR)
	c.Assert(logger.LogLevel(), Equals, loggo.ERROR)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(logger.IsErrorEnabled(), Equals, true)
	c.Assert(logger.IsWarningEnabled(), Equals, false)
	c.Assert(logger.IsInfoEnabled(), Equals, false)
	c.Assert(logger.IsDebugEnabled(), Equals, false)
	c.Assert(logger.IsTraceEnabled(), Equals, false)
	// This is added for completeness, but not really expected to be used.
	logger.SetLogLevel(loggo.CRITICAL)
	c.Assert(logger.LogLevel(), Equals, loggo.CRITICAL)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.CRITICAL)
	c.Assert(logger.IsErrorEnabled(), Equals, false)
	c.Assert(logger.IsWarningEnabled(), Equals, false)
	c.Assert(logger.IsInfoEnabled(), Equals, false)
	c.Assert(logger.IsDebugEnabled(), Equals, false)
	c.Assert(logger.IsTraceEnabled(), Equals, false)
	logger.SetLogLevel(loggo.UNSPECIFIED)
	c.Assert(logger.LogLevel(), Equals, loggo.UNSPECIFIED)
	c.Assert(logger.EffectiveLogLevel(), Equals, loggo.WARNING)
}

func (*loggerSuite) TestLevelsSharedForSameModule(c *C) {
	logger1 := loggo.GetLogger("testing.module")
	logger2 := loggo.GetLogger("testing.module")

	logger1.SetLogLevel(loggo.INFO)
	c.Assert(logger1.IsInfoEnabled(), Equals, true)
	c.Assert(logger2.IsInfoEnabled(), Equals, true)
}

func (*loggerSuite) TestModuleLowered(c *C) {
	logger1 := loggo.GetLogger("TESTING.MODULE")
	logger2 := loggo.GetLogger("Testing")

	c.Assert(logger1.Name(), Equals, "testing.module")
	c.Assert(logger2.Name(), Equals, "testing")
}

func (*loggerSuite) TestLevelsInherited(c *C) {
	root := loggo.GetLogger("")
	first := loggo.GetLogger("first")
	second := loggo.GetLogger("first.second")

	root.SetLogLevel(loggo.ERROR)
	c.Assert(root.LogLevel(), Equals, loggo.ERROR)
	c.Assert(root.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(first.LogLevel(), Equals, loggo.UNSPECIFIED)
	c.Assert(first.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(second.LogLevel(), Equals, loggo.UNSPECIFIED)
	c.Assert(second.EffectiveLogLevel(), Equals, loggo.ERROR)

	first.SetLogLevel(loggo.DEBUG)
	c.Assert(root.LogLevel(), Equals, loggo.ERROR)
	c.Assert(root.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(first.LogLevel(), Equals, loggo.DEBUG)
	c.Assert(first.EffectiveLogLevel(), Equals, loggo.DEBUG)
	c.Assert(second.LogLevel(), Equals, loggo.UNSPECIFIED)
	c.Assert(second.EffectiveLogLevel(), Equals, loggo.DEBUG)

	second.SetLogLevel(loggo.INFO)
	c.Assert(root.LogLevel(), Equals, loggo.ERROR)
	c.Assert(root.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(first.LogLevel(), Equals, loggo.DEBUG)
	c.Assert(first.EffectiveLogLevel(), Equals, loggo.DEBUG)
	c.Assert(second.LogLevel(), Equals, loggo.INFO)
	c.Assert(second.EffectiveLogLevel(), Equals, loggo.INFO)

	first.SetLogLevel(loggo.UNSPECIFIED)
	c.Assert(root.LogLevel(), Equals, loggo.ERROR)
	c.Assert(root.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(first.LogLevel(), Equals, loggo.UNSPECIFIED)
	c.Assert(first.EffectiveLogLevel(), Equals, loggo.ERROR)
	c.Assert(second.LogLevel(), Equals, loggo.INFO)
	c.Assert(second.EffectiveLogLevel(), Equals, loggo.INFO)
}

var parseLevelTests = []struct {
	str   string
	level loggo.Level
	fail  bool
}{{
	str:   "trace",
	level: loggo.TRACE,
}, {
	str:   "TrAce",
	level: loggo.TRACE,
}, {
	str:   "TRACE",
	level: loggo.TRACE,
}, {
	str:   "debug",
	level: loggo.DEBUG,
}, {
	str:   "DEBUG",
	level: loggo.DEBUG,
}, {
	str:   "info",
	level: loggo.INFO,
}, {
	str:   "INFO",
	level: loggo.INFO,
}, {
	str:   "warn",
	level: loggo.WARNING,
}, {
	str:   "WARN",
	level: loggo.WARNING,
}, {
	str:   "warning",
	level: loggo.WARNING,
}, {
	str:   "WARNING",
	level: loggo.WARNING,
}, {
	str:   "error",
	level: loggo.ERROR,
}, {
	str:   "ERROR",
	level: loggo.ERROR,
}, {
	str:   "critical",
	level: loggo.CRITICAL,
}, {
	str:  "not_specified",
	fail: true,
}, {
	str:  "other",
	fail: true,
}, {
	str:  "",
	fail: true,
}}

func (*loggerSuite) TestParseLevel(c *C) {
	for _, test := range parseLevelTests {
		level, ok := loggo.ParseLevel(test.str)
		c.Assert(level, Equals, test.level)
		c.Assert(ok, Equals, !test.fail)
	}
}

var levelStringValueTests = map[loggo.Level]string{
	loggo.UNSPECIFIED: "UNSPECIFIED",
	loggo.DEBUG:       "DEBUG",
	loggo.TRACE:       "TRACE",
	loggo.INFO:        "INFO",
	loggo.WARNING:     "WARNING",
	loggo.ERROR:       "ERROR",
	loggo.CRITICAL:    "CRITICAL",
	loggo.Level(42):   "<unknown>", // other values are unknown
}

func (*loggerSuite) TestLevelStringValue(c *C) {
	for level, str := range levelStringValueTests {
		c.Assert(level.String(), Equals, str)
	}
}

var configureLoggersTests = []struct {
	spec string
	info string
	err  string
}{{
	spec: "",
	info: "<root>=WARNING",
}, {
	spec: "<root>=UNSPECIFIED",
	info: "<root>=WARNING",
}, {
	spec: "<root>=DEBUG",
	info: "<root>=DEBUG",
}, {
	spec: "test.module=debug",
	info: "<root>=WARNING;test.module=DEBUG",
}, {
	spec: "module=info; sub.module=debug; other.module=warning",
	info: "<root>=WARNING;module=INFO;other.module=WARNING;sub.module=DEBUG",
}, {
	spec: "  foo.bar \t\r\n= \t\r\nCRITICAL \t\r\n; \t\r\nfoo \r\t\n = DEBUG",
	info: "<root>=WARNING;foo=DEBUG;foo.bar=CRITICAL",
}, {
	spec: "foo;bar",
	info: "<root>=WARNING",
	err:  `logger specification expected '=', found "foo"`,
}, {
	spec: "=foo",
	info: "<root>=WARNING",
	err:  `logger specification "=foo" has blank name or level`,
}, {
	spec: "foo=",
	info: "<root>=WARNING",
	err:  `logger specification "foo=" has blank name or level`,
}, {
	spec: "=",
	info: "<root>=WARNING",
	err:  `logger specification "=" has blank name or level`,
}, {
	spec: "foo=unknown",
	info: "<root>=WARNING",
	err:  `unknown severity level "unknown"`,
}, {
	// Test that nothing is changed even when the
	// first part of the specification parses ok.
	spec: "module=info; foo=unknown",
	info: "<root>=WARNING",
	err:  `unknown severity level "unknown"`,
}}

func (*loggerSuite) TestConfigureLoggers(c *C) {
	for i, test := range configureLoggersTests {
		c.Logf("test %d: %q", i, test.spec)
		loggo.ResetLoggers()
		err := loggo.ConfigureLoggers(test.spec)
		c.Check(loggo.LoggerInfo(), Equals, test.info)
		if test.err != "" {
			c.Assert(err, ErrorMatches, test.err)
			continue
		}
		c.Assert(err, IsNil)

		// Test that it's idempotent.
		err = loggo.ConfigureLoggers(test.spec)
		c.Assert(err, IsNil)
		c.Assert(loggo.LoggerInfo(), Equals, test.info)

		// Test that calling ConfigureLoggers with the
		// output of LoggerInfo works too.
		err = loggo.ConfigureLoggers(test.info)
		c.Assert(err, IsNil)
		c.Assert(loggo.LoggerInfo(), Equals, test.info)
	}
}

type logwriterSuite struct {
	logger loggo.Logger
	writer *loggo.TestWriter
}

var _ = Suite(&logwriterSuite{})

func (s *logwriterSuite) SetUpTest(c *C) {
	loggo.ResetLoggers()
	loggo.RemoveWriter("default")
	s.writer = &loggo.TestWriter{}
	err := loggo.RegisterWriter("test", s.writer, loggo.TRACE)
	c.Assert(err, IsNil)
	s.logger = loggo.GetLogger("test.writer")
	// Make it so the logger itself writes all messages.
	s.logger.SetLogLevel(loggo.TRACE)
}

func (s *logwriterSuite) TearDownTest(c *C) {
	loggo.ResetWriters()
}

func (s *logwriterSuite) TestLogDoesntLogWeirdLevels(c *C) {
	s.logger.Logf(loggo.UNSPECIFIED, "message")
	c.Assert(s.writer.Log, HasLen, 0)

	s.logger.Logf(loggo.Level(42), "message")
	c.Assert(s.writer.Log, HasLen, 0)

	s.logger.Logf(loggo.CRITICAL+loggo.Level(1), "message")
	c.Assert(s.writer.Log, HasLen, 0)
}

func (s *logwriterSuite) TestMessageFormatting(c *C) {
	s.logger.Logf(loggo.INFO, "some %s included", "formatting")
	c.Assert(s.writer.Log, HasLen, 1)
	c.Assert(s.writer.Log[0].Message, Equals, "some formatting included")
	c.Assert(s.writer.Log[0].Level, Equals, loggo.INFO)
}

func (s *logwriterSuite) BenchmarkLoggingNoWriters(c *C) {
	// No writers
	loggo.RemoveWriter("test")
	for i := 0; i < c.N; i++ {
		s.logger.Warningf("just a simple warning for %d", i)
	}
}

func (s *logwriterSuite) BenchmarkLoggingNoWritersNoFormat(c *C) {
	// No writers
	loggo.RemoveWriter("test")
	for i := 0; i < c.N; i++ {
		s.logger.Warningf("just a simple warning")
	}
}

func (s *logwriterSuite) BenchmarkLoggingTestWriters(c *C) {
	for i := 0; i < c.N; i++ {
		s.logger.Warningf("just a simple warning for %d", i)
	}
	c.Assert(s.writer.Log, HasLen, c.N)
}

func setupTempFileWriter(c *C) (logFile *os.File, cleanup func()) {
	loggo.RemoveWriter("test")
	logFile, err := ioutil.TempFile("", "loggo-test")
	c.Assert(err, IsNil)
	cleanup = func() {
		logFile.Close()
		os.Remove(logFile.Name())
	}
	writer := loggo.NewSimpleWriter(logFile, &loggo.DefaultFormatter{})
	err = loggo.RegisterWriter("testfile", writer, loggo.TRACE)
	c.Assert(err, IsNil)
	return
}

func (s *logwriterSuite) BenchmarkLoggingDiskWriter(c *C) {
	logFile, cleanup := setupTempFileWriter(c)
	defer cleanup()
	msg := "just a simple warning for %d"
	for i := 0; i < c.N; i++ {
		s.logger.Warningf(msg, i)
	}
	offset, err := logFile.Seek(0, os.SEEK_CUR)
	c.Assert(err, IsNil)
	c.Assert((offset > int64(len(msg))*int64(c.N)), Equals, true,
		Commentf("Not enough data was written to the log file."))
}

func (s *logwriterSuite) BenchmarkLoggingDiskWriterNoMessages(c *C) {
	logFile, cleanup := setupTempFileWriter(c)
	defer cleanup()
	// Change the log level
	writer, _, err := loggo.RemoveWriter("testfile")
	c.Assert(err, IsNil)
	loggo.RegisterWriter("testfile", writer, loggo.WARNING)
	msg := "just a simple warning for %d"
	for i := 0; i < c.N; i++ {
		s.logger.Debugf(msg, i)
	}
	offset, err := logFile.Seek(0, os.SEEK_CUR)
	c.Assert(err, IsNil)
	c.Assert(offset, Equals, int64(0),
		Commentf("Data was written to the log file."))
}

func (s *logwriterSuite) BenchmarkLoggingDiskWriterNoMessagesLogLevel(c *C) {
	logFile, cleanup := setupTempFileWriter(c)
	defer cleanup()
	// Change the log level
	s.logger.SetLogLevel(loggo.WARNING)
	msg := "just a simple warning for %d"
	for i := 0; i < c.N; i++ {
		s.logger.Debugf(msg, i)
	}
	offset, err := logFile.Seek(0, os.SEEK_CUR)
	c.Assert(err, IsNil)
	c.Assert(offset, Equals, int64(0),
		Commentf("Data was written to the log file."))
}
