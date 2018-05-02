godel-format-plugin
===================
godel-format-plugin is a gödel plugin that runs Go formatters.

Plugin Tasks
------------
godel-format-plugin provides the following tasks:

* `format`: runs the the format tasks on the files in the project.

Assets
------
godel-format-plugin assets are executables that provide format functionality and also respond to commands that provide information such as the formatter name and verifying formatter configuration.

Writing an asset
----------------
godel-format-plugin provides helper APIs to facilitate writing new assets. More detailed instructions for writing assets are forthcoming. In the meantime, the most effective way to write an asset is to examine the implementation of an existing asset.
