import ast
import json
import sys


def _process_test_diff(file_path):
    with open(file_path) as f:
        diff = json.load(f)

    diff["Packages1"] = _trim_layer_paths(json.dumps(diff["Packages1"]))
    diff["Packages2"] = _trim_layer_paths(json.dumps(diff["Packages2"]))

    with open(file_path, 'w') as f:
        json.dump(diff, f, indent=4)


def _trim_layer_paths(packages):
    packages = ast.literal_eval(packages)
    new_packages = {}
    for package, versions in packages.items():
        versions_to_size = {}
        for path in versions.keys():
            version = versions[path]["Version"]
            size = versions[path]["Size"]
            versions_to_size[version] = size
        new_packages[package] = versions_to_size
    return new_packages


if __name__ == '__main__':
    sys.exit(_process_test_diff(sys.argv[1]))