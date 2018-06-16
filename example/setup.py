"""Example python setup.py."""


from setuptools import setup


setup(
    name="httpy-example",
    version="0.0.1",
    description="example python worker for httpy",
    author="Adam Talsma",
    author_email="adam@talsma.ca",
    url="https://github.com/a-tal/httpy/",
    download_url="https://github.com/a-tal/httpy/",
    classifiers=[
        "Development Status :: 2 - Pre-Alpha",
        "Operating System :: POSIX :: Linux",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Topic :: Software Development",
        "Topic :: Utilities",
    ],
    py_modules=["worker"],
)
