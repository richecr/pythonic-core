import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()


class BinaryDistribution(setuptools.Distribution):
    def has_ext_modules(_):
        return True


setuptools.setup(
    name="pythonic_core",
    version="0.1.0",
    author="Rich Ramalho",
    author_email="richelton14@gmail.com",
    description="This package provides the core functionality for building the query and connecting to the PythonicSQL database.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/richecr/pythonic_core",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: BSD License",
        "Operating System :: OS Independent",
    ],
    include_package_data=True,
    distclass=BinaryDistribution,
)
