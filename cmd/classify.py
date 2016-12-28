import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from sklearn import svm
# def main():
rawsamples=pd.read_csv('sinrtable.csv')
print rawsamples
Nsamples=160
cols=['DISTANCE','SINR']
X=rawsamples.ix[0:Nsamples-1,cols] #.reshape(Nsamples,cols.size)
y=rawsamples.ix[0:Nsamples-1,'TAG'] #.reshape(Nsamples,)

# # take first 50 samples as training TAG, SINR, distance
# training=x.ix [1:50,[0,1,2]]
# # take rest samples 
# samples=x[51:,[0,1,2]]
cl = svm.SVC(kernel='poly')
cl.fit(X,y)

X_test=rawsamples.ix[Nsamples:,cols]
#Xsamples=Xsamples.reshape(Xsamples.size,cols.size)
ypredict=cl.predict(X_test)
ytrue=rawsamples.ix[Nsamples:,'TAG'].values

print ypredict
print ytrue

diff=(ypredict-ytrue)
pdiff=pd.DataFrame(diff)
pdiff[pdiff[0]==1].index

print pdiff[pdiff[0]==1].index
plt.clf()
plt.scatter(X[cols[0]], X[cols[1]], c=y, zorder=10)

# Circle out the test data
# plt.scatter(X_test[cols[0]], X_test[cols[1]],  facecolors='none', zorder=10)

plt.axis('tight')
x_min = X[cols[0]].min()
x_max = X[cols[0]].max()
y_min = X[cols[1]].min()
y_max = X[cols[1]].max()

XX, YY = np.mgrid[x_min:x_max:200j, y_min:y_max:200j]
AZ = cl.decision_function(np.c_[XX.ravel(), YY.ravel()])

# Put the result into a color plot
Z=AZ[:,0]
Z = Z.reshape(XX.shape)
# plt.pcolormesh(XX, YY, Z > 0)
plt.contour(XX, YY, Z, colors=['k', 'k', 'k'], linestyles=['--', '-', '--'],
            levels=[-.5, 0, .5])

# plt.title(kernel)
# plt.show(block=0)

Z=AZ[:,1]
Z = Z.reshape(XX.shape)
# plt.pcolormesh(XX, YY, Z > 0)
plt.contour(XX, YY, Z, colors=['r', 'r', 'r'], linestyles=['--', '-', '--'],
            levels=[-.5, 0, .5])


Z=AZ[:,2]
Z = Z.reshape(XX.shape)
# plt.pcolormesh(XX, YY, Z > 0)
plt.contour(XX, YY, Z, colors=['b', 'b', 'b'], linestyles=['--', '-', '--'],
            levels=[-.5, 0, .5])

# plt.title(kernel)
plt.show()