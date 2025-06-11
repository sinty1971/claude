import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  client: {
    name: '@hey-api/client-fetch',
  },
  input: './openapi.yaml',
  output: './src/api',
  schemas: {
    export: true,
  },
  services: {
    export: true,
    name: '{{name}}Service',
  },
  types: {
    export: true,
  },
});