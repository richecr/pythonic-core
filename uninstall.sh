cd pythonic_core
wheel_file=$(ls dist/*.whl | head -n1); pip uninstall $wheel_file -y
cd ..
rm -rf pythonic_core/