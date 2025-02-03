# TNFSH Course Table Converter

## Prerequisties

- Python 3.8+
- Poetry (recommended) or pip

## Installation


### 1. Clone the repository:

```sh
git clone https://github.com/tobiichi3227/tnfsh-course-table-converter
cd tnfsh-course-table-converter
```

### 2. Install dependencies:

#### Using Poetry (recommended)

If you haven't installed Poetry yet, use the following command:

```sh
curl -sSL https://install.python-poetry.org | python3 -
```

Then, install all dependencies by running:

```sh
poetry install
```

Activate the virtual environment:

```sh
poetry shell
```

#### Using pip

Create and activate a virtual environment:

```sh
python -m venv venv
source venv/bin/activate # On Windows use `venv\Scripts\activate`
```

Then, to install all dependencies, run:

```sh
pip install -r requirements.txt
```

### 3. Run the server:

To start the server, run:

```sh
python server.py # Or `poetry run python server.py` if using Poetry
```

## Contributing

Please see the [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on contributing to this project.

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the [`LICENSE`](./LICENSE) file for details.
