generate:
	gopy pkg -name=pythonic_core -rename=true -author="Rich Ramalho" -email="richelton14@gmail.com" -desc="The unofficial HLTV Python API" -url="https://github.com/richecr/pythonic_core" -output=pythonic_core -vm=`which python` github.com/richecr/pythonic_core github.com/richecr/pythonic_core/lib/pythonic github.com/richecr/pythonic_core/lib/query github.com/richecr/pythonic_core/lib/dialects

install:
	./install.sh

uninstall:
	./uninstall.sh