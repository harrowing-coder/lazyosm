# lazyosm

# What is it? 

Lazyosm is something I've been half attempting to do for a while, it attempts to create a relational model of all the file blocks and the underlying osm pbf file structure to properly utilize assembling features as quickly as possible. This differs drastically from other go implementations of osm data that purely go for memory through of primitive features which doesn't do you a whole lot of good if your node for a way is 100 million primitive features from one another. 

I think what I'm trying to say is this methodology isn't probably the way its intended to be used, throwing all your nodes in a k,v databases and swapping each data structure out to build features is silly (your deserializing nodes just to be written to a file again) but people do it. More importanty the osm file structure I think gives you a few clues on how to do it in a more nuanced manner, but I could be wrong. 

# A small bit of context

Without going deep into osm file structure (I'm not super familiar with it anyway) osm data is effectively a set of file block each containing 8000 primitive features either being densenodes,nodes,ways, or relations. These blocks always have the nodes first, the dense nodes second etc.

### So you end up with file blocks like this:

```
block1: densenode
block2: densenode
block3: densenode
block4: densenode
block5: way
block6: way
block7: relation
```

More importantly hierestically we need to remember that ways are made up of nodes (I think always dense) and relations are made of ways. 

**So knowing what file type is in a certain block has its advantages right out of the gate, but unfortunately the type of block is obsuficated by another message. Most implementations just serialize this without worrying about it, but I have a project that implements a custom pbf reader for situations like this.**

This way of reading only very small crucial parts of the data and leaving everything behind on the first pass allows us to get context that is super valuable. (hopefully) Its also pertaintant to remember that all this requires is gzipping a file block and performing pretty simple operations over a byte array this operation takes nths less time than allocating a huge protobuf. 


# The Context We Get From Preprocessing

```
{Type: IdRange:[0 0] FilePos:[18 183] BufPos:[0 0] Position:0}
{Type:DenseNodes IdRange:[19717967 69668505] FilePos:[200 65350] BufPos:[1138 131580] Position:1}
{Type:DenseNodes IdRange:[69668513 204079406] FilePos:[65367 154737] BufPos:[29742 188425] Position:2}
{Type:DenseNodes IdRange:[204079408 204097234] FilePos:[154754 220744] BufPos:[464 148004] Position:3}
{Type:DenseNodes IdRange:[204097236 204115666] FilePos:[220761 286647] BufPos:[534 148185] Position:4}
{Type:DenseNodes IdRange:[204115667 204133268] FilePos:[286664 353731] BufPos:[563 149501] Position:5}
{Type:DenseNodes IdRange:[204133269 204151453] FilePos:[353748 420277] BufPos:[563 148269] Position:6}
{Type:DenseNodes IdRange:[204151454 204165886] FilePos:[420294 483796] BufPos:[579 146911] Position:7}
{Type:DenseNodes IdRange:[204165888 204181221] FilePos:[483813 548015] BufPos:[460 150120] Position:8}
{Type:DenseNodes IdRange:[204181223 204195465] FilePos:[548032 611805] BufPos:[436 149627] Position:9}
{Type:DenseNodes IdRange:[204195467 204211177] FilePos:[611822 675050] BufPos:[324 149301] Position:10}
{Type:DenseNodes IdRange:[204211178 204223590] FilePos:[675067 743263] BufPos:[404 150754] Position:11}
{Type:DenseNodes IdRange:[204223591 204234913] FilePos:[743280 817350] BufPos:[441 156418] Position:12}
{Type:DenseNodes IdRange:[204234915 204245891] FilePos:[817367 893410] BufPos:[455 159590] Position:13}
{Type:DenseNodes IdRange:[204245892 204254658] FilePos:[893427 967485] BufPos:[519 155947] Position:14}
{Type:DenseNodes IdRange:[204254659 204263175] FilePos:[967502 1042314] BufPos:[530 157281] Position:15}
{Type:DenseNodes IdRange:[204263176 204271790] FilePos:[1042331 1117256] BufPos:[514 157767] Position:16}
{Type:DenseNodes IdRange:[204271791 204280212] FilePos:[1117273 1189913] BufPos:[380 156610] Position:17}
{Type:DenseNodes IdRange:[204280213 204288581] FilePos:[1189930 1263113] BufPos:[516 157553] Position:18}
{Type:DenseNodes IdRange:[204288582 204297011] FilePos:[1263130 1337490] BufPos:[360 158205] Position:19}
{Type:DenseNodes IdRange:[204297013 204308112] FilePos:[1337507 1413287] BufPos:[403 159118] Position:20}
{Type:DenseNodes IdRange:[204308114 204319764] FilePos:[1413304 1485700] BufPos:[283 155990] Position:21}
{Type:DenseNodes IdRange:[204319766 204329976] FilePos:[1485717 1559768] BufPos:[313 161258] Position:22}
{Type:DenseNodes IdRange:[204329978 204340957] FilePos:[1559785 1632809] BufPos:[311 157883] Position:23}
{Type:DenseNodes IdRange:[204340958 204353074] FilePos:[1632826 1707416] BufPos:[581 156598] Position:24}
{Type:DenseNodes IdRange:[204353075 204362055] FilePos:[1707433 1777098] BufPos:[373 154153] Position:25}
{Type:DenseNodes IdRange:[204362056 204370980] FilePos:[1777115 1848647] BufPos:[368 155776] Position:26}
{Type:DenseNodes IdRange:[204370981 204379862] FilePos:[1848664 1922788] BufPos:[597 156839] Position:27}
{Type:DenseNodes IdRange:[204379863 204388649] FilePos:[1922805 1996611] BufPos:[490 157360] Position:28}
{Type:DenseNodes IdRange:[204388650 204397446] FilePos:[1996628 2068947] BufPos:[387 155454] Position:29}
{Type:DenseNodes IdRange:[204397447 204406544] FilePos:[2068964 2142742] BufPos:[455 155142] Position:30}
{Type:DenseNodes IdRange:[204406545 204415776] FilePos:[2142759 2215737] BufPos:[439 154749] Position:31}
{Type:DenseNodes IdRange:[204415777 204426080] FilePos:[2215754 2291133] BufPos:[462 155591] Position:32}
{Type:DenseNodes IdRange:[204426081 204435688] FilePos:[2291150 2365907] BufPos:[451 156230] Position:33}
{Type:DenseNodes IdRange:[204435689 204444257] FilePos:[2365924 2439609] BufPos:[382 155229] Position:34}
{Type:DenseNodes IdRange:[204444258 204453475] FilePos:[2439626 2513781] BufPos:[508 156751] Position:35}
{Type:DenseNodes IdRange:[204453476 204461870] FilePos:[2513798 2586807] BufPos:[439 155729] Position:36}
{Type:DenseNodes IdRange:[204461871 204470234] FilePos:[2586824 2659761] BufPos:[306 156221] Position:37}
{Type:DenseNodes IdRange:[204470235 204478652] FilePos:[2659778 2733957] BufPos:[452 156868] Position:38}
{Type:DenseNodes IdRange:[204478653 204486960] FilePos:[2733974 2808191] BufPos:[499 155951] Position:39}
{Type:DenseNodes IdRange:[204486961 204495298] FilePos:[2808208 2882542] BufPos:[714 157681] Position:40}
{Type:DenseNodes IdRange:[204495299 204504223] FilePos:[2882559 2957043] BufPos:[458 157170] Position:41}
{Type:DenseNodes IdRange:[204504224 204513039] FilePos:[2957060 3030776] BufPos:[570 155717] Position:42}
{Type:DenseNodes IdRange:[204513040 204521827] FilePos:[3030793 3103460] BufPos:[533 156896] Position:43}
{Type:DenseNodes IdRange:[204521828 204531547] FilePos:[3103477 3177314] BufPos:[573 151967] Position:44}
{Type:DenseNodes IdRange:[204531548 204541511] FilePos:[3177331 3252770] BufPos:[707 152591] Position:45}
{Type:DenseNodes IdRange:[204541512 204550293] FilePos:[3252787 3325650] BufPos:[588 152626] Position:46}
{Type:DenseNodes IdRange:[204550294 204559031] FilePos:[3325667 3398304] BufPos:[543 154022] Position:47}
{Type:DenseNodes IdRange:[204559032 204567437] FilePos:[3398321 3471469] BufPos:[476 155233] Position:48}
{Type:DenseNodes IdRange:[204567438 204576000] FilePos:[3471486 3543300] BufPos:[468 152220] Position:49}
{Type:DenseNodes IdRange:[204576001 204584665] FilePos:[3543317 3617428] BufPos:[467 155201] Position:50}
{Type:DenseNodes IdRange:[204584666 204593568] FilePos:[3617445 3691245] BufPos:[570 154582] Position:51}
{Type:DenseNodes IdRange:[204593569 204603211] FilePos:[3691262 3766224] BufPos:[656 155000] Position:52}
{Type:DenseNodes IdRange:[204603212 204611945] FilePos:[3766241 3839910] BufPos:[546 154159] Position:53}
{Type:DenseNodes IdRange:[204611946 204620667] FilePos:[3839927 3913477] BufPos:[476 155288] Position:54}
{Type:DenseNodes IdRange:[204620668 204629709] FilePos:[3913494 3987470] BufPos:[383 156197] Position:55}
{Type:DenseNodes IdRange:[204629710 204638324] FilePos:[3987487 4059806] BufPos:[483 154212] Position:56}
{Type:DenseNodes IdRange:[204638332 204646708] FilePos:[4059823 4131940] BufPos:[350 154408] Position:57}
{Type:DenseNodes IdRange:[204646709 204655176] FilePos:[4131957 4205986] BufPos:[443 155386] Position:58}
{Type:DenseNodes IdRange:[204655177 204663859] FilePos:[4206003 4276359] BufPos:[350 154192] Position:59}
{Type:DenseNodes IdRange:[204663860 204672360] FilePos:[4276376 4349682] BufPos:[392 157312] Position:60}
{Type:DenseNodes IdRange:[204672361 204680744] FilePos:[4349699 4423098] BufPos:[466 156874] Position:61}
{Type:DenseNodes IdRange:[204680745 204689145] FilePos:[4423115 4497383] BufPos:[556 156983] Position:62}
{Type:DenseNodes IdRange:[204689146 204697387] FilePos:[4497400 4568341] BufPos:[392 152729] Position:63}
{Type:DenseNodes IdRange:[204697389 204705819] FilePos:[4568358 4609892] BufPos:[420 129060] Position:64}
{Type:DenseNodes IdRange:[204705820 204714837] FilePos:[4609909 4654319] BufPos:[274 129624] Position:65}
{Type:DenseNodes IdRange:[204714838 204723975] FilePos:[4654336 4697777] BufPos:[474 127074] Position:66}
{Type:DenseNodes IdRange:[204723976 204732948] FilePos:[4697794 4744958] BufPos:[835 127374] Position:67}
{Type:DenseNodes IdRange:[204732949 204741781] FilePos:[4744975 4792354] BufPos:[531 126094] Position:68}
{Type:DenseNodes IdRange:[204741782 204750650] FilePos:[4792371 4842464] BufPos:[390 127581] Position:69}
{Type:DenseNodes IdRange:[204750651 204759488] FilePos:[4842481 4892926] BufPos:[451 127938] Position:70}
{Type:DenseNodes IdRange:[204759489 204767979] FilePos:[4892943 4933364] BufPos:[202 121250] Position:71}
{Type:DenseNodes IdRange:[204767980 204778694] FilePos:[4933381 4981410] BufPos:[342 125481] Position:72}
{Type:DenseNodes IdRange:[204778695 204788094] FilePos:[4981427 5018465] BufPos:[209 117250] Position:73}
{Type:DenseNodes IdRange:[204788095 204797199] FilePos:[5018482 5060559] BufPos:[306 121208] Position:74}
{Type:DenseNodes IdRange:[204797200 204807229] FilePos:[5060576 5104322] BufPos:[314 121329] Position:75}
{Type:DenseNodes IdRange:[204807230 204816892] FilePos:[5104339 5147988] BufPos:[236 123444] Position:76}
{Type:DenseNodes IdRange:[204816893 204826268] FilePos:[5148005 5198775] BufPos:[479 130693] Position:77}
{Type:DenseNodes IdRange:[204826269 204834977] FilePos:[5198792 5239490] BufPos:[421 122835] Position:78}
{Type:DenseNodes IdRange:[204834978 204844468] FilePos:[5239507 5278525] BufPos:[303 120929] Position:79}
{Type:DenseNodes IdRange:[204844469 204853477] FilePos:[5278542 5312982] BufPos:[276 120548] Position:80}
{Type:DenseNodes IdRange:[204853478 204863155] FilePos:[5312999 5347805] BufPos:[293 120525] Position:81}
{Type:DenseNodes IdRange:[204863156 204874505] FilePos:[5347822 5391055] BufPos:[347 126024] Position:82}
{Type:DenseNodes IdRange:[204874506 204884185] FilePos:[5391072 5449644] BufPos:[322 139568] Position:83}
{Type:DenseNodes IdRange:[204884186 204895396] FilePos:[5449661 5501525] BufPos:[480 132660] Position:84}
{Type:DenseNodes IdRange:[204895397 204908261] FilePos:[5501542 5543027] BufPos:[320 123918] Position:85}
{Type:DenseNodes IdRange:[204908262 204920439] FilePos:[5543044 5587581] BufPos:[376 127686] Position:86}
{Type:DenseNodes IdRange:[204920440 204931094] FilePos:[5587598 5638691] BufPos:[435 132406] Position:87}
{Type:DenseNodes IdRange:[204931095 204941992] FilePos:[5638708 5688595] BufPos:[566 128972] Position:88}
{Type:DenseNodes IdRange:[204941993 204952415] FilePos:[5688612 5739519] BufPos:[680 128695] Position:89}
{Type:DenseNodes IdRange:[204952416 204964394] FilePos:[5739536 5789938] BufPos:[555 129145] Position:90}
{Type:DenseNodes IdRange:[204964396 204974406] FilePos:[5789955 5835064] BufPos:[470 131023] Position:91}
{Type:DenseNodes IdRange:[204974407 204983734] FilePos:[5835081 5874087] BufPos:[413 129839] Position:92}
{Type:DenseNodes IdRange:[204983735 204993292] FilePos:[5874104 5913482] BufPos:[351 130360] Position:93}
{Type:DenseNodes IdRange:[204993293 205002347] FilePos:[5913499 5961494] BufPos:[391 130120] Position:94}
{Type:DenseNodes IdRange:[205002348 205011509] FilePos:[5961511 6002811] BufPos:[448 128255] Position:95}
{Type:DenseNodes IdRange:[205011510 205020260] FilePos:[6002828 6052897] BufPos:[340 135392] Position:96}
{Type:DenseNodes IdRange:[205020261 205029068] FilePos:[6052914 6099827] BufPos:[347 129465] Position:97}
{Type:DenseNodes IdRange:[205029069 205038203] FilePos:[6099844 6148063] BufPos:[410 129787] Position:98}
{Type:DenseNodes IdRange:[205038204 205047413] FilePos:[6148080 6190538] BufPos:[349 123896] Position:99}
{Type:DenseNodes IdRange:[205047414 205056691] FilePos:[6190555 6235952] BufPos:[345 128178] Position:100}
{Type:DenseNodes IdRange:[205056692 205066128] FilePos:[6235969 6284134] BufPos:[454 133789] Position:101}
{Type:DenseNodes IdRange:[205066129 205075824] FilePos:[6284151 6321023] BufPos:[349 128415] Position:102}
{Type:DenseNodes IdRange:[205075825 205084882] FilePos:[6321040 6357693] BufPos:[403 129145] Position:103}
{Type:DenseNodes IdRange:[205084883 205094971] FilePos:[6357710 6394168] BufPos:[305 128553] Position:104}
{Type:DenseNodes IdRange:[205094972 205104898] FilePos:[6394185 6432177] BufPos:[324 126187] Position:105}
{Type:DenseNodes IdRange:[205104899 205114338] FilePos:[6432194 6468285] BufPos:[242 129114] Position:106}
{Type:DenseNodes IdRange:[205114339 205123866] FilePos:[6468302 6513729] BufPos:[343 124449] Position:107}
{Type:DenseNodes IdRange:[205123867 205133345] FilePos:[6513746 6558778] BufPos:[564 128788] Position:108}
{Type:DenseNodes IdRange:[205133346 205142531] FilePos:[6558795 6606423] BufPos:[456 129444] Position:109}
{Type:DenseNodes IdRange:[205142535 205151555] FilePos:[6606440 6653659] BufPos:[397 128780] Position:110}
{Type:DenseNodes IdRange:[205151556 205160477] FilePos:[6653676 6698105] BufPos:[364 125250] Position:111}
{Type:DenseNodes IdRange:[205160478 205169586] FilePos:[6698122 6753730] BufPos:[370 134445] Position:112}
{Type:DenseNodes IdRange:[205169587 205178686] FilePos:[6753747 6805375] BufPos:[411 129496] Position:113}
{Type:DenseNodes IdRange:[205178687 205188179] FilePos:[6805392 6859650] BufPos:[402 133025] Position:114}
{Type:DenseNodes IdRange:[205188180 205198842] FilePos:[6859667 6918294] BufPos:[454 135220] Position:115}
{Type:DenseNodes IdRange:[205198843 205214499] FilePos:[6918311 6973053] BufPos:[347 133116] Position:116}
{Type:DenseNodes IdRange:[205214500 205236524] FilePos:[6973070 7021105] BufPos:[537 130084] Position:117}
{Type:DenseNodes IdRange:[205236530 205265515] FilePos:[7021122 7066451] BufPos:[409 125316] Position:118}
{Type:DenseNodes IdRange:[205265516 205281622] FilePos:[7066468 7109691] BufPos:[286 124721] Position:119}
{Type:DenseNodes IdRange:[205281623 205317969] FilePos:[7109708 7154280] BufPos:[347 124444] Position:120}
{Type:DenseNodes IdRange:[205317970 205334452] FilePos:[7154297 7196053] BufPos:[296 125006] Position:121}
{Type:DenseNodes IdRange:[205334453 205344354] FilePos:[7196070 7238388] BufPos:[338 125544] Position:122}
{Type:DenseNodes IdRange:[205344355 205353859] FilePos:[7238405 7284303] BufPos:[459 127536] Position:123}
{Type:DenseNodes IdRange:[205353860 205363635] FilePos:[7284320 7331919] BufPos:[673 128874] Position:124}
{Type:DenseNodes IdRange:[205363636 205373053] FilePos:[7331936 7374703] BufPos:[363 130113] Position:125}
{Type:DenseNodes IdRange:[205373054 205383968] FilePos:[7374720 7415219] BufPos:[287 128535] Position:126}
{Type:DenseNodes IdRange:[205383969 205393113] FilePos:[7415236 7455318] BufPos:[300 128474] Position:127}
{Type:DenseNodes IdRange:[205393114 205403202] FilePos:[7455335 7498361] BufPos:[444 129034] Position:128}
{Type:DenseNodes IdRange:[205403203 205413757] FilePos:[7498378 7539404] BufPos:[427 128785] Position:129}
{Type:DenseNodes IdRange:[205413758 205426606] FilePos:[7539421 7579995] BufPos:[398 128552] Position:130}
{Type:DenseNodes IdRange:[205426607 205436489] FilePos:[7580012 7621542] BufPos:[341 128573] Position:131}
{Type:DenseNodes IdRange:[205436490 205446764] FilePos:[7621559 7663632] BufPos:[330 127967] Position:132}
{Type:DenseNodes IdRange:[205446765 205457452] FilePos:[7663649 7706530] BufPos:[283 128585] Position:133}
{Type:DenseNodes IdRange:[205457453 205467806] FilePos:[7706547 7749059] BufPos:[310 127755] Position:134}
{Type:DenseNodes IdRange:[205467807 205477626] FilePos:[7749076 7792207] BufPos:[374 128427] Position:135}
{Type:DenseNodes IdRange:[205477627 205487706] FilePos:[7792224 7835327] BufPos:[344 129008] Position:136}
{Type:DenseNodes IdRange:[205487707 221582776] FilePos:[7835344 7880367] BufPos:[503 126315] Position:137}
{Type:DenseNodes IdRange:[221582777 303892003] FilePos:[7880384 7950386] BufPos:[1594 127209] Position:138}
{Type:DenseNodes IdRange:[303892004 331913946] FilePos:[7950403 7998306] BufPos:[1329 100475] Position:139}
{Type:DenseNodes IdRange:[331913947 356558594] FilePos:[7998323 8167376] BufPos:[149697 369587] Position:140}
{Type:DenseNodes IdRange:[356558595 366480497] FilePos:[8167393 8230871] BufPos:[6537 122078] Position:141}
{Type:DenseNodes IdRange:[366480610 444205286] FilePos:[8230888 8312303] BufPos:[38156 166058] Position:142}
{Type:DenseNodes IdRange:[444205520 469689553] FilePos:[8312320 8358192] BufPos:[723 100265] Position:143}
{Type:DenseNodes IdRange:[469689554 474705919] FilePos:[8358209 8396375] BufPos:[359 92402] Position:144}
{Type:DenseNodes IdRange:[474705922 489207429] FilePos:[8396392 8435234] BufPos:[237 90462] Position:145}
{Type:DenseNodes IdRange:[489207430 566266612] FilePos:[8435251 8484178] BufPos:[1934 99261] Position:146}
{Type:DenseNodes IdRange:[566266614 566545086] FilePos:[8484195 8525372] BufPos:[359 93883] Position:147}
{Type:DenseNodes IdRange:[566545087 570168517] FilePos:[8525389 8567466] BufPos:[437 95184] Position:148}
{Type:DenseNodes IdRange:[570168519 617926383] FilePos:[8567483 8611546] BufPos:[11416 126503] Position:149}
{Type:DenseNodes IdRange:[617926384 617953316] FilePos:[8611563 8649190] BufPos:[200 93702] Position:150}
{Type:DenseNodes IdRange:[617953317 634786594] FilePos:[8649207 8694814] BufPos:[792 97468] Position:151}
{Type:DenseNodes IdRange:[634786602 657511150] FilePos:[8694831 8739379] BufPos:[227 98080] Position:152}
{Type:DenseNodes IdRange:[657511152 660983896] FilePos:[8739396 8774219] BufPos:[83 93362] Position:153}
{Type:DenseNodes IdRange:[660983898 661913552] FilePos:[8774236 8809873] BufPos:[127 91383] Position:154}
{Type:DenseNodes IdRange:[661913553 662010121] FilePos:[8809890 8848193] BufPos:[72 92636] Position:155}
{Type:DenseNodes IdRange:[662010128 702536220] FilePos:[8848210 8889215] BufPos:[1641 95066] Position:156}
{Type:DenseNodes IdRange:[702536222 745786926] FilePos:[8889232 8931385] BufPos:[1622 95178] Position:157}
{Type:DenseNodes IdRange:[745786928 778886466] FilePos:[8931402 8976928] BufPos:[4219 102943] Position:158}
{Type:DenseNodes IdRange:[778886467 810264273] FilePos:[8976945 9016860] BufPos:[888 95326] Position:159}
{Type:DenseNodes IdRange:[810264274 837136452] FilePos:[9016877 9054860] BufPos:[196 93934] Position:160}
{Type:DenseNodes IdRange:[837136454 865607067] FilePos:[9054877 9111979] BufPos:[432 113324] Position:161}
{Type:DenseNodes IdRange:[865607068 865667579] FilePos:[9111996 9179168] BufPos:[155 121717] Position:162}
{Type:DenseNodes IdRange:[865667589 977202111] FilePos:[9179185 9236504] BufPos:[803 110563] Position:163}
{Type:DenseNodes IdRange:[977202113 1157140554] FilePos:[9236521 9295538] BufPos:[2294 116214] Position:164}
{Type:DenseNodes IdRange:[1157140556 1173705358] FilePos:[9295555 9348867] BufPos:[716 107452] Position:165}
{Type:DenseNodes IdRange:[1173705361 1187500258] FilePos:[9348884 9412882] BufPos:[829 117054] Position:166}
{Type:DenseNodes IdRange:[1187500269 1209162694] FilePos:[9412899 9473476] BufPos:[1522 109795] Position:167}
{Type:DenseNodes IdRange:[1209162706 1211077531] FilePos:[9473493 9538926] BufPos:[1196 117064] Position:168}
{Type:DenseNodes IdRange:[1211077533 1214360025] FilePos:[9538943 9606527] BufPos:[2194 122715] Position:169}
{Type:DenseNodes IdRange:[1214360028 1241497198] FilePos:[9606544 9669095] BufPos:[3883 116580] Position:170}
{Type:DenseNodes IdRange:[1241497199 1266117597] FilePos:[9669112 9721807] BufPos:[557 105524] Position:171}
{Type:DenseNodes IdRange:[1266117603 1300602299] FilePos:[9721824 9781439] BufPos:[1594 112464] Position:172}
{Type:DenseNodes IdRange:[1300602304 1309699468] FilePos:[9781456 9836657] BufPos:[475 110232] Position:173}
{Type:DenseNodes IdRange:[1309699469 1309713517] FilePos:[9836674 9878026] BufPos:[97 102911] Position:174}
{Type:DenseNodes IdRange:[1309713520 1309725542] FilePos:[9878043 9914966] BufPos:[71 96399] Position:175}
{Type:DenseNodes IdRange:[1309725544 1347473294] FilePos:[9914983 9963734] BufPos:[1407 103843] Position:176}
{Type:DenseNodes IdRange:[1347473295 1485394416] FilePos:[9963751 10009100] BufPos:[1323 99940] Position:177}
{Type:DenseNodes IdRange:[1485394417 1544077196] FilePos:[10009117 10056441] BufPos:[1901 104206] Position:178}
{Type:DenseNodes IdRange:[1544077200 1617022125] FilePos:[10056458 10108836] BufPos:[4852 108461] Position:179}
{Type:DenseNodes IdRange:[1617022127 1682357862] FilePos:[10108853 10156847] BufPos:[3835 102755] Position:180}
{Type:DenseNodes IdRange:[1682357864 1721326127] FilePos:[10156864 10209298] BufPos:[2241 106781] Position:181}
{Type:DenseNodes IdRange:[1721326128 1733688025] FilePos:[10209315 10263505] BufPos:[964 109727] Position:182}
{Type:DenseNodes IdRange:[1733688026 1822160734] FilePos:[10263522 10315182] BufPos:[2443 108760] Position:183}
{Type:DenseNodes IdRange:[1822160735 1840594769] FilePos:[10315199 10362358] BufPos:[1210 102357] Position:184}
{Type:DenseNodes IdRange:[1840594771 1872936631] FilePos:[10362375 10404540] BufPos:[549 97243] Position:185}
{Type:DenseNodes IdRange:[1872936632 1975229871] FilePos:[10404557 10454718] BufPos:[1499 104221] Position:186}
{Type:DenseNodes IdRange:[1975229874 2006136095] FilePos:[10454735 10496697] BufPos:[2736 100480] Position:187}
{Type:DenseNodes IdRange:[2006136107 2116699853] FilePos:[10496714 10545518] BufPos:[5461 107839] Position:188}
{Type:DenseNodes IdRange:[2116699854 2116711363] FilePos:[10545535 10581597] BufPos:[50 97268] Position:189}
{Type:DenseNodes IdRange:[2116711364 2116729297] FilePos:[10581614 10625435] BufPos:[97 107531] Position:190}
{Type:DenseNodes IdRange:[2116729299 2116744612] FilePos:[10625452 10663507] BufPos:[80 99184] Position:191}
{Type:DenseNodes IdRange:[2116744618 2116755159] FilePos:[10663524 10702976] BufPos:[73 103161] Position:192}
{Type:DenseNodes IdRange:[2116755160 2116769833] FilePos:[10702993 10735168] BufPos:[65 89275] Position:193}
{Type:DenseNodes IdRange:[2116769834 2116781112] FilePos:[10735185 10766448] BufPos:[37 89207] Position:194}
{Type:DenseNodes IdRange:[2116781113 2147700596] FilePos:[10766465 10807364] BufPos:[807 97462] Position:195}
{Type:DenseNodes IdRange:[2147700597 2194573380] FilePos:[10807381 10847868] BufPos:[433 96584] Position:196}
{Type:DenseNodes IdRange:[2194573381 2224043924] FilePos:[10847885 10889379] BufPos:[3194 103027] Position:197}
{Type:DenseNodes IdRange:[2224043925 2291989481] FilePos:[10889396 10931275] BufPos:[1870 97278] Position:198}
{Type:DenseNodes IdRange:[2291989482 2380002427] FilePos:[10931292 10978156] BufPos:[1922 103571] Position:199}
{Type:DenseNodes IdRange:[2380002428 2395661474] FilePos:[10978173 11020969] BufPos:[1638 98293] Position:200}
{Type:DenseNodes IdRange:[2395661475 2415506983] FilePos:[11020986 11054147] BufPos:[871 90335] Position:201}
{Type:DenseNodes IdRange:[2415506984 2419408679] FilePos:[11054164 11089572] BufPos:[387 93078] Position:202}
{Type:DenseNodes IdRange:[2419408680 2434487976] FilePos:[11089589 11123941] BufPos:[1576 90192] Position:203}
{Type:DenseNodes IdRange:[2434487978 2467230538] FilePos:[11123958 11164031] BufPos:[845 95565] Position:204}
{Type:DenseNodes IdRange:[2467230539 2478364458] FilePos:[11164048 11206294] BufPos:[1546 99553] Position:205}
{Type:DenseNodes IdRange:[2478364459 2485894264] FilePos:[11206311 11247161] BufPos:[523 97704] Position:206}
{Type:DenseNodes IdRange:[2485894266 2491765883] FilePos:[11247178 11290357] BufPos:[492 100911] Position:207}
{Type:DenseNodes IdRange:[2491765884 2509636237] FilePos:[11290374 11330548] BufPos:[2248 96184] Position:208}
{Type:DenseNodes IdRange:[2509636238 2539160901] FilePos:[11330565 11368335] BufPos:[2080 93505] Position:209}
{Type:DenseNodes IdRange:[2539160903 2603969603] FilePos:[11368352 11412790] BufPos:[4254 101216] Position:210}
{Type:DenseNodes IdRange:[2603969604 2631778430] FilePos:[11412807 11455537] BufPos:[2115 97015] Position:211}
{Type:DenseNodes IdRange:[2631778431 2642676863] FilePos:[11455554 11490640] BufPos:[337 89482] Position:212}
{Type:DenseNodes IdRange:[2642676865 2699338343] FilePos:[11490657 11532846] BufPos:[2029 99489] Position:213}
{Type:DenseNodes IdRange:[2699338344 2727140421] FilePos:[11532863 11569081] BufPos:[740 92312] Position:214}
{Type:DenseNodes IdRange:[2727140422 2751055075] FilePos:[11569098 11606723] BufPos:[617 93917] Position:215}
{Type:DenseNodes IdRange:[2751055076 2769578612] FilePos:[11606740 11641661] BufPos:[358 90992] Position:216}
{Type:DenseNodes IdRange:[2769578614 2805216010] FilePos:[11641678 11682786] BufPos:[934 96681] Position:217}
{Type:DenseNodes IdRange:[2805216012 2835991861] FilePos:[11682803 11723691] BufPos:[1444 96358] Position:218}
{Type:DenseNodes IdRange:[2835991862 2916044994] FilePos:[11723708 11762046] BufPos:[1670 93959] Position:219}
{Type:DenseNodes IdRange:[2916044995 2942533047] FilePos:[11762063 11798725] BufPos:[1296 92621] Position:220}
{Type:DenseNodes IdRange:[2942533048 3003894147] FilePos:[11798742 11835811] BufPos:[1685 93509] Position:221}
{Type:DenseNodes IdRange:[3003894148 3090763600] FilePos:[11835828 11874829] BufPos:[1828 94787] Position:222}
{Type:DenseNodes IdRange:[3090763601 3122850137] FilePos:[11874846 11909937] BufPos:[945 91855] Position:223}
{Type:DenseNodes IdRange:[3122850138 3165304029] FilePos:[11909954 11948555] BufPos:[2119 93885] Position:224}
{Type:DenseNodes IdRange:[3165304030 3195049050] FilePos:[11948572 11983836] BufPos:[1684 92291] Position:225}
{Type:DenseNodes IdRange:[3195049051 3252778984] FilePos:[11983853 12025782] BufPos:[4549 98312] Position:226}
{Type:DenseNodes IdRange:[3252778985 3331278671] FilePos:[12025799 12063839] BufPos:[1777 93560] Position:227}
{Type:DenseNodes IdRange:[3331278672 3338898404] FilePos:[12063856 12091259] BufPos:[135 86622] Position:228}
{Type:DenseNodes IdRange:[3338898405 3432634453] FilePos:[12091276 12131259] BufPos:[2782 96428] Position:229}
{Type:DenseNodes IdRange:[3432634454 3522813147] FilePos:[12131276 12171002] BufPos:[1265 94913] Position:230}
{Type:DenseNodes IdRange:[3522813148 3599743680] FilePos:[12171019 12208496] BufPos:[1492 92772] Position:231}
{Type:DenseNodes IdRange:[3599743681 3649812927] FilePos:[12208513 12247303] BufPos:[2569 96146] Position:232}
{Type:DenseNodes IdRange:[3649812928 3709827979] FilePos:[12247320 12287810] BufPos:[1869 96243] Position:233}
{Type:DenseNodes IdRange:[3709827980 3755278338] FilePos:[12287827 12329236] BufPos:[2243 97091] Position:234}
{Type:DenseNodes IdRange:[3755278339 3772343829] FilePos:[12329253 12368179] BufPos:[1788 94486] Position:235}
{Type:DenseNodes IdRange:[3772343830 3806129306] FilePos:[12368196 12403824] BufPos:[1972 91800] Position:236}
{Type:DenseNodes IdRange:[3806129307 3814828525] FilePos:[12403841 12435452] BufPos:[537 87932] Position:237}
{Type:DenseNodes IdRange:[3814828526 3839866694] FilePos:[12435469 12473393] BufPos:[1417 92950] Position:238}
{Type:DenseNodes IdRange:[3839866695 3916785643] FilePos:[12473410 12510073] BufPos:[1089 93101] Position:239}
{Type:DenseNodes IdRange:[3916785644 4083380593] FilePos:[12510090 12551466] BufPos:[2655 96171] Position:240}
{Type:DenseNodes IdRange:[4083380594 4211116154] FilePos:[12551483 12591281] BufPos:[3147 95249] Position:241}
{Type:DenseNodes IdRange:[4211116155 4233924834] FilePos:[12591298 12626934] BufPos:[974 92032] Position:242}
{Type:DenseNodes IdRange:[4233924835 4244400778] FilePos:[12626951 12656571] BufPos:[321 88935] Position:243}
{Type:DenseNodes IdRange:[4244400779 4341311291] FilePos:[12656588 12698630] BufPos:[3265 98270] Position:244}
{Type:DenseNodes IdRange:[4341311292 4382292507] FilePos:[12698647 12739798] BufPos:[1559 95889] Position:245}
{Type:DenseNodes IdRange:[4382292508 4427213008] FilePos:[12739815 12781628] BufPos:[5629 100020] Position:246}
{Type:DenseNodes IdRange:[4427250030 4456456925] FilePos:[12781645 12815680] BufPos:[1064 91138] Position:247}
{Type:DenseNodes IdRange:[4456456926 4475186187] FilePos:[12815697 12850378] BufPos:[874 91842] Position:248}
{Type:DenseNodes IdRange:[4475186188 4494804905] FilePos:[12850395 12885860] BufPos:[1019 92682] Position:249}
{Type:DenseNodes IdRange:[4494804906 4542773562] FilePos:[12885877 12925241] BufPos:[2556 96158] Position:250}
{Type:DenseNodes IdRange:[4542773563 4598598770] FilePos:[12925258 12965700] BufPos:[4022 96763] Position:251}
{Type:DenseNodes IdRange:[4598598771 4673130629] FilePos:[12965717 13005278] BufPos:[3174 95109] Position:252}
{Type:DenseNodes IdRange:[4673130630 4702464715] FilePos:[13005295 13043741] BufPos:[2772 94827] Position:253}
{Type:DenseNodes IdRange:[4702464716 4752875396] FilePos:[13043758 13077847] BufPos:[1929 91151] Position:254}
{Type:DenseNodes IdRange:[4752875397 4779317150] FilePos:[13077864 13113634] BufPos:[397 90980] Position:255}
{Type:DenseNodes IdRange:[4779317151 4912292477] FilePos:[13113651 13155165] BufPos:[3649 97344] Position:256}
{Type:DenseNodes IdRange:[4912292478 4974247459] FilePos:[13155182 13198249] BufPos:[7922 101668] Position:257}
{Type:DenseNodes IdRange:[4974247460 4995599083] FilePos:[13198266 13236752] BufPos:[2423 95221] Position:258}
{Type:DenseNodes IdRange:[4995599084 5026513850] FilePos:[13236769 13272100] BufPos:[1338 91839] Position:259}
{Type:DenseNodes IdRange:[5026513851 5049837123] FilePos:[13272117 13302378] BufPos:[462 89246] Position:260}
{Type:DenseNodes IdRange:[5049837124 5146027416] FilePos:[13302395 13340337] BufPos:[2092 93115] Position:261}
{Type:DenseNodes IdRange:[5146027417 5219864614] FilePos:[13340354 13379527] BufPos:[1089 94059] Position:262}
{Type:DenseNodes IdRange:[5219864615 5250510189] FilePos:[13379544 13418175] BufPos:[1421 93953] Position:263}
{Type:DenseNodes IdRange:[5250510190 5264952455] FilePos:[13418192 13455330] BufPos:[1395 93616] Position:264}
{Type:DenseNodes IdRange:[5264952456 5276550485] FilePos:[13455347 13490066] BufPos:[1072 90768] Position:265}
{Type:DenseNodes IdRange:[5276550486 5294831470] FilePos:[13490083 13524231] BufPos:[1505 90800] Position:266}
{Type:DenseNodes IdRange:[5294831471 5314965493] FilePos:[13524248 13557642] BufPos:[958 89718] Position:267}
{Type:DenseNodes IdRange:[5314965494 5333832402] FilePos:[13557659 13591828] BufPos:[1620 90574] Position:268}
{Type:DenseNodes IdRange:[5333832403 5347188815] FilePos:[13591845 13627703] BufPos:[865 92280] Position:269}
{Type:DenseNodes IdRange:[5347188816 5359679205] FilePos:[13627720 13661038] BufPos:[769 90817] Position:270}
{Type:DenseNodes IdRange:[5359679206 5363930840] FilePos:[13661054 13669873] BufPos:[64 22934] Position:271}
{Type:Ways IdRange:[0 0] FilePos:[13669890 13989632] BufPos:[124178 805956] Position:272}
{Type:Ways IdRange:[0 0] FilePos:[13989649 14339357] BufPos:[130729 819981] Position:273}
{Type:Ways IdRange:[0 0] FilePos:[14339374 14658717] BufPos:[121194 745259] Position:274}
{Type:Ways IdRange:[0 0] FilePos:[14658734 14935073] BufPos:[122914 762575] Position:275}
{Type:Ways IdRange:[0 0] FilePos:[14935090 15168957] BufPos:[112974 684426] Position:276}
{Type:Ways IdRange:[0 0] FilePos:[15168974 15488233] BufPos:[138418 824200] Position:277}
{Type:Ways IdRange:[0 0] FilePos:[15488250 15829886] BufPos:[133901 831184] Position:278}
{Type:Ways IdRange:[0 0] FilePos:[15829903 16421065] BufPos:[49142 981650] Position:279}
{Type:Ways IdRange:[0 0] FilePos:[16421082 16739254] BufPos:[60170 662509] Position:280}
{Type:Ways IdRange:[0 0] FilePos:[16739271 16924682] BufPos:[18235 489641] Position:281}
{Type:Ways IdRange:[0 0] FilePos:[16924699 17208092] BufPos:[48679 652036] Position:282}
{Type:Ways IdRange:[0 0] FilePos:[17208109 17462908] BufPos:[37350 599750] Position:283}
{Type:Ways IdRange:[0 0] FilePos:[17462925 17675736] BufPos:[37087 561496] Position:284}
{Type:Ways IdRange:[0 0] FilePos:[17675753 17860170] BufPos:[36784 511561] Position:285}
{Type:Ways IdRange:[0 0] FilePos:[17860187 18053718] BufPos:[35614 520886] Position:286}
{Type:Ways IdRange:[0 0] FilePos:[18053735 18200191] BufPos:[30370 491124] Position:287}
{Type:Ways IdRange:[0 0] FilePos:[18200208 18250116] BufPos:[9507 167325] Position:288}
{Type:Relations IdRange:[0 0] FilePos:[18250133 18432208] BufPos:[56647 406936] Position:289}
```

