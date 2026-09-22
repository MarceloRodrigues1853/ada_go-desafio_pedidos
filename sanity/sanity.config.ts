import {defineConfig} from 'sanity'
import {structureTool} from 'sanity/structure'
import {visionTool} from '@sanity/vision'
import {schemaTypes} from './schemaTypes'

const projectId = process.env.SANITY_STUDIO_PROJECT_ID

if (!projectId) throw new Error('SANITY_STUDIO_PROJECT_ID não configurada')

export default defineConfig({
  name: 'open-coach',
  title: 'OpenCoach Knowledge Studio',
  projectId,
  dataset: process.env.SANITY_STUDIO_DATASET || 'production',
  plugins: [structureTool(), visionTool()],
  schema: {types: schemaTypes},
})
