.POSIX:
.SUFFIXES:

CARGO = cargo
RM = rm
INSTALL = install
SCDOC = scdoc
PREFIX = /usr/local
BINDIR = bin
MANDIR = share/man

all: mata

mata:
	$(CARGO) build --release

clean:
	$(RM) -rf target result

lint:
	$(CARGO) clippy

test:
	$(CARGO) test

install:
	$(INSTALL) -d \
		$(DESTDIR)$(PREFIX)/$(BINDIR)/

	$(INSTALL) -pm 0755 target/release/mata $(DESTDIR)$(PREFIX)/$(BINDIR)/

uninstall:
	$(RM) -f \
		$(DESTDIR)$(PREFIX)/$(BINDIR)/mata

.PHONY: all mata clean install uninstall
