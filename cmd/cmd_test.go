package cmd

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestAsUserCommand(t *testing.T) {
	Convey("When the AsUser function is called with a string", t, func() {
		cmd := AsUser("gfrey", "do something")
		Convey("Then the resulting command must be executed as the given user", func() {
			So(cmd.user, ShouldEqual, "gfrey")
		})
	})

	Convey("Given a ShellCommand", t, func() {
		sc := &ShellCommand{Command: "do something"}
		Convey("When the AsUser function is called with it", func() {
			cmd := AsUser("gfrey", sc)
			Convey("Then the resulting command must be executed as the given user", func() {
				So(cmd.user, ShouldEqual, "gfrey")
			})
		})

		Convey("Given the user is already set", func() {
			sc.user = "gfrey"
			Convey("When the AsUser function is called with it", func() {
				f := func() { AsUser("gfrey", sc) }
				Convey("Then the AsUser functin must panic", func() {
					So(f, ShouldPanicWith, `nesting "AsUser" calls not supported`)
				})
			})
		})
	})

	Convey("Given a FileCommand", t, func() {
		fc := &FileCommand{Path: "tmpf", Content: "foobar"}
		Convey("When the AsUser function is called with it", func() {
			f := func() { AsUser("gfrey", fc) }
			Convey("Then the AsUser functin must panic", func() {
				So(f, ShouldPanicWith, `type "*cmd.FileCommand" not supported`)
			})
		})
	})
}

func TestUpdatePackagesCommand(t *testing.T) {
	Convey("When the UpdatePackages command is called", t, func() {
		cmd := UpdatePackages()
		Convey("Then the command must contain apt-get update", func() {
			So(cmd.Shell(), ShouldContainSubstring, "apt-get update")
		})
		Convey("Then the command must contain apt-get upgrade", func() {
			So(cmd.Shell(), ShouldContainSubstring, "apt-get upgrade")
		})
	})
}

func TestInstallPackagesCommand(t *testing.T) {
	Convey("When the InstallPackages command is called for a package foo", t, func() {
		c := InstallPackages("foo")
		Convey("Then the result should contain the foo package", func() {
			So(c.Shell(), ShouldContainSubstring, "foo")
		})
	})

	Convey("When the InstallPackages command is called for packages foo and bar", t, func() {
		c := InstallPackages("foo", "bar")
		Convey("Then the result should contain the both packages", func() {
			So(c.Shell(), ShouldContainSubstring, "foo")
			So(c.Shell(), ShouldContainSubstring, "bar")
		})
	})
}

func TestAndCommand(t *testing.T) {
	Convey("When the And command is called for a command foo", t, func() {
		c := And("foo")
		Convey("Then the result should only contain the foo command", func() {
			So(c.Shell(), ShouldEqual, "foo")
		})
	})

	Convey("When the And command is called for commands foo and bar", t, func() {
		c := And("foo", "bar")
		Convey("Then the result should contain the combined commands", func() {
			So(c.Shell(), ShouldEqual, "{ foo && bar; }")
		})
	})

	Convey("When the And command is called for mixed commands", t, func() {
		c := And(&ShellCommand{Command: "foo"}, "bar")
		Convey("Then the result should contain the combined commands", func() {
			So(c.Shell(), ShouldEqual, "{ foo && bar; }")
		})
	})

	Convey("When the And command is called for mixed commands where one has a user set", t, func() {
		f := func() { And(&ShellCommand{Command: "foo", user: "gfrey"}, "bar") }
		Convey("Then And function must panic", func() {
			So(f, ShouldPanicWith, "AsUser not supported in nested commands")
		})
	})

	Convey("When the And command is called for mixed commands where one is not a ShellCommand", t, func() {
		f := func() { And(&FileCommand{Path: "foo"}, "bar") }
		Convey("Then And function must panic", func() {
			So(f, ShouldPanicWith, `type "*cmd.FileCommand" not supported`)
		})
	})
}

func TestOrCommand(t *testing.T) {
	Convey("When the Or command is called for a command foo", t, func() {
		c := Or("foo")
		Convey("Then the result should only contain the foo command", func() {
			So(c.Shell(), ShouldEqual, "foo")
		})
	})

	Convey("When the Or command is called for commands foo and bar", t, func() {
		c := Or("foo", "bar")
		Convey("Then the result should contain the combined commands", func() {
			So(c.Shell(), ShouldEqual, "{ foo || bar; }")
		})
	})

	Convey("When the Or command is called for mixed commands", t, func() {
		c := Or(&ShellCommand{Command: "foo"}, "bar")
		Convey("Then the result should contain the combined commands", func() {
			So(c.Shell(), ShouldEqual, "{ foo || bar; }")
		})
	})

	Convey("When the Or command is called for mixed commands where one has a user set", t, func() {
		f := func() { Or(&ShellCommand{Command: "foo", user: "gfrey"}, "bar") }
		Convey("Then Or function must panic", func() {
			So(f, ShouldPanicWith, "AsUser not supported in nested commands")
		})
	})

	Convey("When the Or command is called for mixed commands where one is not a ShellCommand", t, func() {
		f := func() { Or(&FileCommand{Path: "foo"}, "bar") }
		Convey("Then Or function must panic", func() {
			So(f, ShouldPanicWith, `type "*cmd.FileCommand" not supported`)
		})
	})
}

func TestMkdirCommand(t *testing.T) {
	Convey("When the Mkdir command is called without a path", t, func() {
		Convey("Then the function panics", func() {
			f := func() { Mkdir("", "", 0) }
			So(f, ShouldPanicWith, "empty path given to mkdir")
		})
	})

	Convey("Given the path '/tmp/foo'", t, func() {
		path := "/tmp/foo"
		Convey("When neither owner nor mode are set", func() {
			owner := ""
			var mode os.FileMode = 0
			Convey("Then the mkdir command won't set owner or permissions", func() {
				c := Mkdir(path, owner, mode)
				So(c.Shell(), ShouldEqual, "mkdir -p /tmp/foo")
			})
		})

		Convey("When the owner is set", func() {
			owner := "gfrey"
			var mode os.FileMode = 0
			Convey("Then the mkdir command will change the owner", func() {
				c := Mkdir(path, owner, mode)
				So(c.Shell(), ShouldContainSubstring, "chown gfrey /tmp/foo")
			})
		})

		Convey("When the mode is set", func() {
			owner := ""
			var mode os.FileMode = 0755
			Convey("Then the mkdir command will change the permissions", func() {
				c := Mkdir(path, owner, mode)
				So(c.Shell(), ShouldContainSubstring, "chmod 755 /tmp/foo")
			})
		})

		Convey("When both owner and mode are set", func() {
			owner := "gfrey"
			var mode os.FileMode = 0755
			Convey("Then the mkdir command will change owner and permissions", func() {
				c := Mkdir(path, owner, mode)
				So(c.Shell(), ShouldContainSubstring, "chown gfrey /tmp/foo")
				So(c.Shell(), ShouldContainSubstring, "chmod 755 /tmp/foo")
			})
		})
	})
}

func TestIfCommand(t *testing.T) {
	Convey("When the If command is called without a test", t, func() {
		f := func() { If("", "") }
		Convey("Then the function panics", func() {
			So(f, ShouldPanicWith, "empty test given")
		})
	})

	Convey("When the If command is called with a test", t, func() {
		test := "-d /tmp"
		Convey("When the If command is called without a command", func() {
			f := func() { If(test, "") }
			Convey("Then the function panics", func() {
				So(f, ShouldPanicWith, "empty command given")
			})
		})
	})

	Convey("Given the test '-d /tmp'", t, func() {
		test := "-d /tmp"
		Convey("Given the command 'echo \"true\"'", func() {
			cmd := "echo \"true\""
			Convey("Then the resulting command will contain both", func() {
				c := If(test, cmd)
				So(c.Shell(), ShouldContainSubstring, test)
				So(c.Shell(), ShouldContainSubstring, cmd)
			})
		})

		Convey("Given a ShellCommand", func() {
			cmd := &ShellCommand{Command: `echo "true"`}
			Convey("Then the resulting command will contain both", func() {
				c := If(test, cmd)
				So(c.Shell(), ShouldContainSubstring, test)
				So(c.Shell(), ShouldContainSubstring, `echo "true"`)
			})
		})

		Convey("Given a FileCommand", func() {
			cmd := &FileCommand{Path: "/tmpf"}
			f := func() { If(test, cmd) }
			Convey("Then the function panics", func() {
				So(f, ShouldPanicWith, `type "*cmd.FileCommand" not supported`)
			})
		})
	})
}

func TestIfNotCommand(t *testing.T) {
	Convey("When the IfNot command is called without a test", t, func() {
		f := func() { IfNot("", "") }
		Convey("Then the function panics", func() {
			So(f, ShouldPanicWith, "empty test given")
		})
	})

	Convey("When the IfNot command is called with a test", t, func() {
		test := "-d /tmp"
		Convey("When the IfNot command is called without a command", func() {
			f := func() { IfNot(test, "") }
			Convey("Then the function panics", func() {
				So(f, ShouldPanicWith, "empty command given")
			})
		})
	})

	Convey("Given the test '-d /tmp'", t, func() {
		test := "-d /tmp"
		Convey("Given the command 'echo \"true\"'", func() {
			cmd := "echo \"true\""
			Convey("When the IfNot command is called with those", func() {
				c := IfNot(test, cmd)
				Convey("Then the result must contain both", func() {
					So(c.Shell(), ShouldContainSubstring, test)
					So(c.Shell(), ShouldContainSubstring, cmd)
				})
			})
		})

		Convey("Given a ShellCommand", func() {
			cmd := &ShellCommand{Command: `echo "true"`}
			Convey("Then the resulting command will contain both", func() {
				c := IfNot(test, cmd)
				So(c.Shell(), ShouldContainSubstring, test)
				So(c.Shell(), ShouldContainSubstring, `echo "true"`)
			})
		})

		Convey("Given a FileCommand", func() {
			cmd := &FileCommand{Path: "/tmpf"}
			f := func() { IfNot(test, cmd) }
			Convey("Then the function panics", func() {
				So(f, ShouldPanicWith, `type "*cmd.FileCommand" not supported`)
			})
		})
	})
}
