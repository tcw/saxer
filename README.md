# SAXER

## Description

After working with XML for years there has been one *nix tool that I've been missing.
A xml exploration tool that can handle a subset of xpath queries for very large files.
This is an attempt at creating such a tool.

This tool is in alpha state!

## Install

    go get github.com/tcw/saxer

## Usage

    usage: saxer [<flags>] <query> [<file>]

    Flags:
         --help         Show context-sensitive help (also try --help-long and --help-man).
     -i, --inner        Inner-xml of selected element (default false)
     -n, --count        Number of matches (default false)
     -m, --meta         Get query meta data - linenumbers and path of matches (default false)
     -f, --firstN=0     First n matches (default (0 = all matches))
     -u, --unescape     Unescape html escape tokens (&lt; &gt; ...)
     -s, --case         Turn on case insensitivity
     -o, --omit-ns      Omit namespace in tag-name matches
     -c, --contains     Maching of tag-name and attributes is executed by contains (not equals)
     -w, --wrap         Wrap result in Xml tag
         --tag-buf=4    Size of element tag buffer in KB - tag size
         --cont-buf=4   Size of content buffer in MB - returned elements size
         --profile-cpu  Profile parser
         --version      Show application version.

    Args:
     <query>   Sax query expression
     [<file>]  xml-file


### Simple Usage

TODO

## Examples


### example.xml

    <cars>
    	<car vin="wp031" man="Volvo">
    		<color>blue</color>
    		<xs:doors>4</xs:doors>
    		<engine nr="001">
    			<Fuel>Gasoline</Fuel>
    		</engine>
    	</car>
    	<car vin="wp032" man="Volvo">
    		<color>red</color>
    		<xs:doors>2</xs:doors>
    		<engine nr="002">
    			<Fuel>Diesel</Fuel>
    		</engine>
    	</car>
    	<car vin="wp033" man="Saab">
    		<color>yellow</color>
    		<xs:doors>4</xs:doors>
    		<engine nr="003">
    			<Fuel>Diesel</Fuel>
    		</engine>
    	</car>
    	<info>&lt;some-xml>data&lt;/some-xml></info>
    </cars>

Command:

    saxer engine example.xml

    Returns:
    <engine nr="001">
      <fuel>Gasoline</fuel>
    </engine>
    <engine nr="002">
      <fuel>Diesel</fuel>
    </engine>
    <engine nr="002">
      <fuel>Diesel</fuel>
    </engine>

Command:

    saxer engine?nr=001 example.xml

    Returns:
    <engine nr="001">
        <Fuel>Gasoline</Fuel>
      </engine>

Command:

    saxer ?man=Volvo example.xml

    Returns:
    <car vin="wp031" man="Volvo">
      <color>blue</color>
      <xs:doors>4</xs:doors>
      <engine nr="001">
        <Fuel>Gasoline</Fuel>
      </engine>
    </car>
    <car vin="wp032" man="Volvo">
      <color>red</color>
      <xs:doors>2</xs:doors>
      <engine nr="002">
        <Fuel>Diesel</Fuel>
      </engine>
    </car>

Command:

    saxer "?vin=wp031&man=Volvo" example.xml

    Returns:
    <car vin="wp031" man="Volvo">
      <color>blue</color>
      <xs:doors>4</xs:doors>
      <engine nr="001">
        <Fuel>Gasoline</Fuel>
      </engine>
    </car>

Command:

    saxer -i Fuel example.xml

    Returns:
     Gasoline
     Diesel
     Diesel

Command:

    saxer -n Fuel example.xml

    Returns:
     3


Command:

    saxer -m engine example.xml

    Returns:
     5-7    cars/car/engine
     12-14    cars/car/engine
     19-21    cars/car/engine

Command:

    saxer -f 2 Fuel example.xml

    Returns:
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>

Command:

    saxer -u info example.xml

    Returns:
    <info><some-xml>data</some-xml></info>

Command:

    saxer -s fuel example.xml

    Returns:
    <Fuel>Gasoline</Fuel>
    <Fuel>Diesel</Fuel>
    <Fuel>Diesel</Fuel>

Command:

    saxer -o doors example.xml

    Returns:
    <xs:doors>4</xs:doors>
    <xs:doors>2</xs:doors>
    <xs:doors>4</xs:doors>


Command:

    saxer -c or example.xml

    Returns:
    <color>blue</color>
    <xs:doors>4</xs:doors>
    <color>red</color>
    <xs:doors>2</xs:doors>
    <color>yellow</color>
    <xs:doors>4</xs:doors>

Command:

    saxer -w Fuel example.xml

    Returns:
    <saxer-result>
      <Fuel>Gasoline</Fuel>
      <Fuel>Diesel</Fuel>
      <Fuel>Diesel</Fuel>
    </saxer-result>




