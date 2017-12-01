const CleanWebpackPlugin = require("clean-webpack-plugin");
const ExtractTextPlugin = require("extract-text-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const ManifestPlugin = require("webpack-manifest-plugin");
const MinifyPlugin = require("babel-minify-webpack-plugin");
const path = require("path");
const webpack = require("webpack");

const extractSass = new ExtractTextPlugin({
  filename: "css/[name].[hash].css"
});

module.exports = {
  context: path.resolve(__dirname, "src"),
  entry: "./js/app.js",
  output: {
    filename: "js/[name].[hash].js",
    path: path.resolve(__dirname, "assets/static")
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader",
          options: {
            presets: ["es2016"]
          }
        }
      },
      {
        test: /\.scss$/,
        use: extractSass.extract({
          use: [
            {
              loader: "css-loader"
            },
            {
              loader: "postcss-loader",
              options: {
                ident: "postcss",
                plugins: loader => [
                  require("postcss-import")({ root: loader.resourcePath }),
                  require("autoprefixer")(),
                  require("cssnano")()
                ]
              }
            },
            {
              loader: "sass-loader"
            }
          ],
          // use style-loader in development
          fallback: "style-loader"
        })
      }
    ]
  },
  plugins: [
    new CleanWebpackPlugin(["assets"]),
    new ManifestPlugin({
      fileName: "../manifest.json",
      publicPath: "assets/static/"
    }),
    new HtmlWebpackPlugin({
      filename: "../index.html",
      template: "index.html"
    }),
    new MinifyPlugin(),
    new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
    extractSass
  ]
};
