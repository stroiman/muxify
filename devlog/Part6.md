# September 12-14th 2024 - Create an executable

tldr; Create an executable binary that reads configuration from a file, and I am
now starting to use the tool.

---

Finally, some progress after a long break.

After the last session, I had the functionality to start, resume a session (i.e.
recreate missing windows/panes, and rerun their commands), but no means to
actually launch the code.

During these days, I managed to

 1. Create the functionality to parse configuration from a .yaml file
 2. Locate a configuration file in a proper location in the file system.
 3. Bind it together with a proper `main()` function that actually does the work.

## 1. Read configuration from a yaml.

This wasn't much of a problem in itself; I was familiar with the idea of having
metadata tags on struct members that can be used to control
serialization/deserialization.

The first iteration would work with string constants representing the
configuration.

## 2. Locate the configuration file.

I wanted this to be covered by tests, as the is a little (not much) complexity
to the discovery of configuration files. The standard library has an
`testing/fstest` package, but that is not enough in itself to represent a full
fake file system.

This took some discovery in learning what this package does, and add
functionality on top to that allows the test to express the fixture as a list of
files in a global file system, and a set of environment variables.

## 3. Wrap it into an executable

For this to happen, first I needed some mocking functionality. The main entry
point has the responsibility of parsing options, loading the configuration, and
calling the correct function based on that. 

So for the test for the main entry point logic
- A configuration file in a mocked file system
- A set of environment variables
- A set of arguments for the executable

The verification of the test is that it calls the correct application logic to
launch a project, passing the correct project struct. To do this, an interface
was created for the boundary between these two layers; and I decided to use
[gomock](https://github.com/uber-go/mock) to help veryifying the expected call.

Now it was time to create an executable, and here it did show it had been some
years since I last worked with the language. I had used the package name
`muxify` for my code in the root folder. But this needs to be named `main` for
the project to be compiled to an executable; rather than a library.
