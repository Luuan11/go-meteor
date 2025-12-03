const GAME_CONSTANTS = {
  CACHE_DURATION: 30000,
  MIN_SCORE_INTERVAL: 5000,
  RECAPTCHA_TIMEOUT: 5000,
  MIN_NAME_LENGTH: 2,
  MAX_NAME_LENGTH: 15,
  MAX_SCORE: 999999,
  MIN_SCORE: 0,
  LEADERBOARD_SIZE: 10,
  AUTO_REFRESH_INTERVAL: 30000,
  NOTIFICATION_DURATION: 3000,
};

const API_CONFIG = {
  BASE_URL: 'https://go-meteor.vercel.app/api/leaderboard',
  RECAPTCHA_SITE_KEY: '6LdWFgwsAAAAAAMzR76ilX1OUF56FtKjU2yOlvcG',
  VERSION: '0.2.0'
};

const VALIDATION = {
  NAME_PATTERN: /^[a-zA-Z0-9\s\-_]+$/,
  NAME_MIN: 2,
  NAME_MAX: 15
};

if (typeof module !== 'undefined' && module.exports) {
  module.exports = { GAME_CONSTANTS, API_CONFIG, VALIDATION };
}
