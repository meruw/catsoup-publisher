"""Read-only JSON adapter. Never complete, review or approve a delivery."""
import importlib.util
import json
import pathlib
import sys

root = pathlib.Path(__file__).resolve().parents[3]
spec = importlib.util.spec_from_file_location("catsoup_contract", root / "contracts/delivery/delivery.py")
contract = importlib.util.module_from_spec(spec)
spec.loader.exec_module(contract)
errors, warnings, publishable, _ = contract.check(sys.argv[1])
print(json.dumps({"errors": errors, "warnings": warnings, "publishable": publishable}))
