gopy pkg -name=pythonic_core -rename=true -author="Rich Ramalho" -email="richelton14@gmail.com" -desc="This package provides the core functionality for building the query and connecting to the PythonicSQL database." -url="https://github.com/richecr/pythonic_core" -output=pythonic_core -vm=`which python` github.com/richecr/pythonic_core github.com/richecr/pythonic_core/lib/pythonic github.com/richecr/pythonic_core/lib/query github.com/richecr/pythonic_core/lib/dialects

cp README_lib.md ./pythonic_core/README.md
cp setup.py ./pythonic_core

cd pythonic_core/
python setup.py bdist_wheel
python -m build

wheel_file=$(ls dist/*.whl | head -n1); pip install $wheel_file