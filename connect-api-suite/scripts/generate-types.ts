import { generateZodClientFromOpenAPI } from 'openapi-zod-client';
import { writeFileSync } from 'fs';
import { z } from 'zod';

// Read OpenAPI URL from CLI args or env, fallback to local backend server
const argv = process.argv.slice(2);
let openApiUrl = process.env.OPENAPI_URL;
for (let i = 0; i < argv.length; i++) {
  const arg = argv[i];
  if (!arg.startsWith('-') && !openApiUrl) {
    openApiUrl = arg;
    break;
  }
  if (arg === '--url' || arg === '-u') {
    openApiUrl = argv[i + 1];
    break;
  }
  const match = arg.match(/^--url=(.+)$/);
  if (match) {
    openApiUrl = match[1];
    break;
  }
}
openApiUrl = openApiUrl || 'http://localhost:8080/swagger/doc.json';

console.log(`📡 Fetching OpenAPI spec from: ${openApiUrl}`);

try {
  const res = await fetch(openApiUrl);
  
  if (!res.ok) {
    throw new Error(`Failed to fetch OpenAPI spec: ${res.status} ${res.statusText}`);
  }
  
  const openApiDoc = await res.json();

  console.log(`✅ OpenAPI spec fetched successfully`);
  console.log(`📝 API Title: ${openApiDoc.info?.title || 'Unknown'}`);
  console.log(`📝 API Version: ${openApiDoc.info?.version || 'Unknown'}`);
  console.log(`🔨 Generating TypeScript types and Zod schemas...`);

  await generateZodClientFromOpenAPI({
    openApiDoc,
    distPath: './src/types/generated/',
    handlebars: hbs,
    options: {
      groupStrategy: 'tag-file',
      exportTypes: true,
      withAlias: true,
      apiClientName: 'api',
      exportSchemas: true,
      withDocs: true,
      shouldExportAllSchemas: true,
      withDeprecatedEndpoints: false,
    }
  });

  console.log(`✨ Types generated successfully in ./src/types/generated/`);
  console.log(`📦 Generated files include:`);
  console.log(`   - TypeScript types for all endpoints`);
  console.log(`   - Zod schemas for validation`);
  console.log(`   - API client functions`);
  console.log(`🎉 Done! You can now import from '@/types/generated'`);
} catch (error) {
  console.error(`❌ Error generating types:`, error);
  process.exit(1);
}
