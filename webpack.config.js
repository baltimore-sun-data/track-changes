const CleanWebpackPlugin = require("clean-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const ManifestPlugin = require("webpack-manifest-plugin");
const MinifyPlugin = require("babel-minify-webpack-plugin");
const path = require("path");
const webpack = require("webpack");

module.exports = {
  entry: "./src/js/app.js",
  output: {
    filename: "js/[name].[hash].js",
    path: path.resolve(__dirname, "assets")
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
      }
    ]
  },
  plugins: [
    new CleanWebpackPlugin(["assets"]),
    new ManifestPlugin(),
    new HtmlWebpackPlugin({
      filename: "index.html",
      template: "src/index.html"
    }),
    new MinifyPlugin(),
    new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/)
  ]
};
