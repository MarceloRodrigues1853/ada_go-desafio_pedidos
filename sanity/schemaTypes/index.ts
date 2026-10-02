import {apiEndpoint} from './apiEndpoint'
import {architectureDecision} from './architectureDecision'
import {businessRule} from './businessRule'
import {documentationClaim} from './documentationClaim'
import {domainEvent} from './domainEvent'

export const schemaTypes = [
  businessRule,
  architectureDecision,
  domainEvent,
  apiEndpoint,
  documentationClaim,
]
