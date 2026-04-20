from setuptools import find_packages, setup

setup(
    name="vd_stats",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "rich>=13.0",
        "paramiko>=3.0",
    ],
    entry_points={
        "console_scripts": [
            "vd_stats=vd_stats.main:main",
        ],
    },
    python_requires=">=3.8",
)
