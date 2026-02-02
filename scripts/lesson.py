#!/usr/bin/env python3

import sys


def main(argv) -> int:
    if len(argv) < 2:
        print("Usage: python scripts/lesson.py <argument>")
        print("len(argv):", len(argv))
        return 1

    argument = argv[1] if len(argv) == 2 else "too few args"
    print(f"Argument received: {argument}")
    print("len(argv):", len(argv))
    return 0  
  
if __name__ == "__main__":
    sys.exit(main(sys.argv))