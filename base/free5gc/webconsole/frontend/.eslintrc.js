module.exports = {
    "env": {
        "browser": true,
        "es2021": true
    },
  "root": true,
    "extends": [
      "airbnb-typescript/base",
        "eslint:recommended",
        "plugin:react/recommended",
        "plugin:@typescript-eslint/recommended",
        "prettier"
    ],
    "overrides": [
    ],
    "parser": "@typescript-eslint/parser",
    "parserOptions": {
        "ecmaVersion": "latest",
      "sourceType": "module",
      "project": "./tsconfig.json"
    },
    "plugins": [
        "react",
      "@typescript-eslint",
      "import"
    ],
    "rules": {
      "@typescript-eslint/no-non-null-assertion": 0,
      "@typescript-eslint/no-explicit-any": 0,
      "import/extensions": "off"
    }
}
