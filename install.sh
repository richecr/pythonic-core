gopy pkg -name=pythonicsql -rename=true -author="Rich Ramalho" -email="richelton14@gmail.com" -desc="The unofficial HLTV Python API" -url="https://github.com/richecr/pythonicsqlgo" -output=pythonicsql -vm=`which python` github.com/richecr/pythonicsqlgo github.com/richecr/pythonicsqlgo/lib/pythonic github.com/richecr/pythonicsqlgo/lib/query github.com/richecr/pythonicsqlgo/lib/dialects

cd pythonicsql/
python setup.py bdist_wheel
python -m build

wheel_file=$(ls dist/*.whl | head -n1); pip install $wheel_file