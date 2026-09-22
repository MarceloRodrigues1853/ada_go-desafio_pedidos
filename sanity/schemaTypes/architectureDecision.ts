import {defineField, defineType} from 'sanity'

export const architectureDecision = defineType({
  name: 'architectureDecision',
  title: 'Decisão arquitetural',
  type: 'document',
  fields: [
    defineField({name: 'title', title: 'Título', type: 'string', validation: (rule) => rule.required()}),
    defineField({name: 'context', title: 'Contexto', type: 'text', rows: 4}),
    defineField({name: 'decision', title: 'Decisão', type: 'text', rows: 4, validation: (rule) => rule.required()}),
    defineField({name: 'consequences', title: 'Consequências', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'components', title: 'Componentes afetados', type: 'array', of: [{type: 'string'}]}),
    defineField({name: 'source', title: 'Fonte', type: 'string'}),
    defineField({name: 'reviewedAt', title: 'Revisada em', type: 'datetime'}),
  ],
})
