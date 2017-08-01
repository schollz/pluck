import lxml.html
import time
t = time.time()
hparser = lxml.html.HTMLParser(encoding='utf-8')
doc = lxml.etree.parse(open('test.txt', 'r'), hparser)
print(doc.xpath('//a[@rel="bookmark"]')[0].text_content())
print(time.time() - t)
