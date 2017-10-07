import time
import json

import lxml.html


def parseHTML():
    hparser = lxml.html.HTMLParser(encoding='utf-8')
    doc = lxml.etree.parse(open('test.txt', 'r'), hparser)
    result = {}
    result['songs'] = []
    for i in doc.xpath('//a[@rel="bookmark"]'):
        result['songs'].append(i.text_content())

    for i in doc.xpath('//title'):
        result['title'] = i.text_content()

    result['options'] = []
    for j, i in enumerate(doc.xpath('//option[@class="level-0"]')):
        result['options'].append(i.text_content())
        if j == 3:
            break
    return result

if __name__ == '__main__':
    print(json.dumps(parseHTML(), indent=2))
    import timeit
    print(1/(timeit.timeit("parseHTML()",
                        setup="from __main__ import parseHTML", number=100) / 100))
