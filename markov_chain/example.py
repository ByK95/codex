import random
from collections import defaultdict

class MarkovChain:
    def __init__(self):
        self.model = defaultdict(list)

    def train(self, data):
        for i in range(len(data) - 1):
            self.model[data[i]].append(data[i + 1])
            print(self.model)

    def generate(self, start, length=10):
        result = [start]
        for _ in range(length - 1):
            next_states = self.model.get(result[-1])
            if not next_states:
                break
            result.append(random.choice(next_states))
        return result

# Example usage
chain = MarkovChain()
sequence = ["walk", "run", "jump", "walk", "run", "walk", "jump"]
chain.train(sequence)
print(chain.generate("walk", 5))
