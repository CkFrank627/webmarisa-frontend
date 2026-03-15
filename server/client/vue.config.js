module.exports = {
  devServer: {
    proxy: {
      '^/marisa': { target: 'http://localhost:3000', changeOrigin: true },
      '^/Add': { target: 'http://localhost:3000', changeOrigin: true },
      '^/Reply': { target: 'http://localhost:3000', changeOrigin: true },
      '^/Forget': { target: 'http://localhost:3000', changeOrigin: true },
      '^/Status': { target: 'http://localhost:3000', changeOrigin: true },
    }
  }
}
