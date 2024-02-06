module.exports = {
  testEnvironment: 'jsdom',
  transform: {
    "\\.[jt]sx?$": "babel-jest",
  },
  transformIgnorePatterns: [
    "node_modules/(?!(p-queue|p-retry|p-timeout|is-network-error)/)"
  ],
};