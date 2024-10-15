export enum ErrorType {
  NO_STATE_COOKIE = 'no_state_cookie',
  STATE_MISMATCH = 'state_mismatch',
  TOKEN_EXCHANGE = 'token_exchange',
  DATABASE_ERROR = 'database_error',
  MISSING_ID_TOKEN = 'missing_id_token',
  INVALID_ID_TOKEN = 'invalid_id_token',
  MISSING_JWT_PRIVATE_KEY = 'missing_jwt_private_key',
  ERROR_ADDING_TOKEN = 'error_adding_token',
  ERROR_GENERATING_JWT = 'error_generating_jwt',
  ERROR_ID_TOKEN_CLAIMS = 'error_id_token_claims',
  ERROR_ADDING_USER = 'error_adding_user',
}

export const ERROR_MESSAGES: Record<ErrorType, string> = {
  [ErrorType.NO_STATE_COOKIE]: 'Authentication failed. Please try again later.',
  [ErrorType.STATE_MISMATCH]: 'Authentication failed. Please try again later.',
  [ErrorType.TOKEN_EXCHANGE]: 'Unable to complete login. Please try again later.',
  [ErrorType.DATABASE_ERROR]: 'Unable to complete login. Please try again later.',
  [ErrorType.MISSING_ID_TOKEN]: 'Authentication failed. Please try again later.',
  [ErrorType.INVALID_ID_TOKEN]: 'Authentication failed. Please try again later.',
  [ErrorType.ERROR_ID_TOKEN_CLAIMS]: 'Authentication failed. Please try again later.',
  [ErrorType.ERROR_ADDING_USER]: 'Authentication failed. Please try again later.',
  [ErrorType.ERROR_GENERATING_JWT]: 'Authentication failed. Please try again later.',
  [ErrorType.MISSING_JWT_PRIVATE_KEY]: 'Authentication failed. Please try again later.',
  [ErrorType.ERROR_ADDING_TOKEN]: 'Authentication failed. Please try again later.',
};

export const DEFAULT_ERROR_MESSAGE = 'An unexpected error occurred. Please try again.';
